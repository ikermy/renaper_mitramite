package telegram

import (
	"renaper_mitramite/internal/application"

	tb "gopkg.in/telebot.v4"
)

type Adapter struct {
	bot     *tb.Bot
	usecase *application.Bot
}

func NewAdapter(bot *tb.Bot, usecase *application.Bot) *Adapter {
	return &Adapter{bot: bot, usecase: usecase}
}

func (a *Adapter) Register() {
	a.bot.Handle("/start", func(c tb.Context) error {
		reply, err := a.usecase.HandleCommand(int(c.Sender().ID), "start")
		if err != nil {
			return err
		}
		return a.sendReply(c, reply)
	})

	a.bot.Handle("/help", func(c tb.Context) error {
		reply, err := a.usecase.HandleCommand(int(c.Sender().ID), "help")
		if err != nil {
			return err
		}
		return a.sendReply(c, reply)
	})

	a.bot.Handle("/check", func(c tb.Context) error {
		reply, err := a.usecase.HandleCommand(int(c.Sender().ID), "check")
		if err != nil {
			return err
		}
		return a.sendReply(c, reply)
	})

	a.bot.Handle(tb.OnCallback, func(c tb.Context) error {
		reply, err := a.usecase.HandleCallback(int(c.Sender().ID), c.Callback().Data)
		if err != nil {
			return err
		}
		return a.sendReply(c, reply)
	})

	a.bot.Handle(tb.OnText, func(c tb.Context) error {
		reply, err := a.usecase.HandleText(int(c.Sender().ID), c.Text())
		if err != nil {
			return err
		}
		return a.sendReply(c, reply)
	})
}

func (a *Adapter) sendReply(c tb.Context, reply application.Reply) error {
	if reply.Text == "" {
		return nil
	}
	if reply.Markup == nil {
		return c.Send(reply.Text)
	}
	return c.Send(reply.Text, a.toMarkup(reply.Markup))
}

func (a *Adapter) toMarkup(markup *application.Markup) *tb.ReplyMarkup {
	buttons := make([][]tb.InlineButton, 0, len(markup.InlineKeyboard))
	for _, row := range markup.InlineKeyboard {
		buttonsRow := make([]tb.InlineButton, 0, len(row))
		for _, button := range row {
			buttonsRow = append(buttonsRow, tb.InlineButton{Text: button.Text, Data: button.Data})
		}
		buttons = append(buttons, buttonsRow)
	}
	return &tb.ReplyMarkup{InlineKeyboard: buttons}
}
