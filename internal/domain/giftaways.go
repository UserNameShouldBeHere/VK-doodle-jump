package domain

import (
	"time"

	"github.com/tarantool/go-tarantool/v2/datetime"
)

type Gift struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Photo       string `json:"photo"`
	Count       int    `json:"count"`
}

type GiftawayData struct {
	Id          int       `json:"id"`
	Description string    `json:"description"`
	Details     string    `json:"details"`
	From        time.Time `json:"from"`
	To          time.Time `json:"to"`
}

type GiftawayDataT struct {
	Id          int               `json:"id"`
	Description string            `json:"description"`
	Details     string            `json:"details"`
	From        datetime.Datetime `json:"from"`
	To          datetime.Datetime `json:"to"`
}

type GiftawayT struct {
	Info  GiftawayDataT `json:"info"`
	Gifts []Gift        `json:"gifts"`
}

type Giftaway struct {
	Info  GiftawayData `json:"info"`
	Gifts []Gift       `json:"gifts"`
}
