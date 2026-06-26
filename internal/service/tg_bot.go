package service

import (
	"context"
	"fmt"
	"home-library/internal/models"
	"home-library/internal/repository"
	"log"
	"net/http"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TgBotService struct {
	bot        *tgbotapi.BotAPI
	bookRepo   *repository.BookRepository
	loanRepo   *repository.LoanRepository
	userRepo   *repository.UserRepository
	isbnSvc    *ISBNService
	barcodeSvc *BarcodeService
}

func NewTgBotService(token string, br *repository.BookRepository, lr *repository.LoanRepository, ur *repository.UserRepository, isbn *ISBNService, bar *BarcodeService) (*TgBotService, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &TgBotService{
		bot:        bot,
		bookRepo:   br,
		loanRepo:   lr,
		userRepo:   ur,
		isbnSvc:    isbn,
		barcodeSvc: bar,
	}, nil
}

func (s *TgBotService) Start() {
	log.Printf("Tg bot @%s успешно запущен!", s.bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := s.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if len(update.Message.Photo) > 0 {
			go s.handlePhoto(update.Message)
			continue
		}

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				args := update.Message.CommandArguments()
				if args != "" {
					go s.handleTelegramBinding(update.Message, args)
				} else {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет! Я бот твоей домашней библиотеки.\n\n📸 Отправь мне фото штрих-кода книги, чтобы добавить её.\n📋 Пиши /debtors, чтобы увидеть список должников.")
					s.bot.Send(msg)
				}
			case "debtors":
				go s.handleDebtors(update.Message)
			default:
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда 🤔")
				s.bot.Send(msg)
			}
		}
	}
}

func (s *TgBotService) handleDebtors(msg *tgbotapi.Message) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := s.userRepo.GetByTelegramChatID(ctx, msg.Chat.ID)
	if err != nil {
		s.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "❌ Ты не авторизован в системе. Сделай привязку через личный кабинет."))
		return
	}

	loans, err := s.loanRepo.GetActiveLoans(ctx, user.ID)
	if err != nil {
		s.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Ошибка получения списка должников ❌"))
		return
	}

	if len(loans) == 0 {
		s.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Ура! Все книги на полках, должников нет 🎉"))
		return
	}

	responseText := "📚 Книги у друзей:\n\n"
	for _, loan := range loans {
		responseText += fmt.Sprintf("👤 %s\n📅 Вернуть до: %s\n\n", loan.BorrowerName, loan.DueDate.Format("02.01.2006"))
	}

	s.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, responseText))
}

func (s *TgBotService) handlePhoto(msg *tgbotapi.Message) {
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	user, err := s.userRepo.GetByTelegramChatID(ctx, msg.Chat.ID)
	if err != nil {
		s.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "❌ Ты не авторизован в системе. Сделай привязку через личный кабинет."))
		return
	}

	s.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Сканирую фотографию... 🔍"))

	photo := msg.Photo[len(msg.Photo)-1]

	fileURL, err := s.bot.GetFileDirectURL(photo.FileID)
	if err != nil {
		s.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Не удалось скачать фото из Telegram ❌"))
		return
	}

	// Скачиваем картинку в память
	resp, err := http.Get(fileURL)
	if err != nil {
		s.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Ошибка скачивания файла ❌"))
		return
	}
	defer resp.Body.Close()

	isbnCode, err := s.barcodeSvc.ScanISBN(resp.Body)
	if err != nil {
		s.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Не удалось распознать штрих-код. Сделайте более четкое фото под прямым углом 📷"))
		return
	}

	bookInfo, err := s.isbnSvc.FetchBookInfo(ctx, isbnCode)
	if err != nil {
		s.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Штрих-код считан: `%s`,\nно книга не найдена во внешних базах 🤷‍♂️", isbnCode)))
		return
	}

	newBook := models.Book{
		OwnerID:     user.ID,
		Title:       bookInfo.Title,
		Authors:     bookInfo.Authors,
		ISBN:        isbnCode,
		Description: bookInfo.Description,
	}

	_, err = s.bookRepo.Create(ctx, newBook)
	if err != nil {
		s.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, fmt.Errorf("Ошибка сохранения книги в базу: %w", err).Error()))
		return
	}

	successMsg := fmt.Sprintf("🎉 Книга успешно добавлена в библиотеку!\n\n📖 *Название:* %s\n✍️ *Авторы:* %v\n🆔 *ISBN:* %s", bookInfo.Title, bookInfo.Authors, isbnCode)
	reply := tgbotapi.NewMessage(msg.Chat.ID, successMsg)
	reply.ParseMode = "Markdown"
	s.bot.Send(reply)
}

func (s *TgBotService) handleTelegramBinding(msg *tgbotapi.Message, token string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := s.userRepo.LinkTelegramByToken(ctx, token, msg.Chat.ID)
	if err != nil {
		s.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "❌ Ошибка авторизации: ссылка устарела или токен неверен."))
		return
	}

	successMsg := fmt.Sprintf("🎉 Успех, %s! Твой Telegram успешно связан с домашней библиотекой. Теперь ты можешь добавлять книги со своего аккаунта!", user.Name)
	s.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, successMsg))
}
