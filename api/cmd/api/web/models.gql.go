// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package web

type ClapInput struct {
	SpeakerID *string `json:"speakerId"`
	ConfaID   *string `json:"confaId"`
	TalkID    *string `json:"talkId"`
}

type Claps struct {
	Value     int `json:"value"`
	UserValue int `json:"userValue"`
}

type Confa struct {
	ID      string `json:"id"`
	OwnerID string `json:"ownerId"`
	Handle  string `json:"handle"`
}

type ConfaInput struct {
	ID      *string `json:"id"`
	OwnerID *string `json:"ownerId"`
	Handle  *string `json:"handle"`
}

type Confas struct {
	Items    []*Confa `json:"items"`
	Limit    int      `json:"limit"`
	NextFrom string   `json:"nextFrom"`
}

type Talk struct {
	ID        string `json:"id"`
	OwnerID   string `json:"ownerId"`
	SpeakerID string `json:"speakerId"`
	ConfaID   string `json:"confaId"`
	RoomID    string `json:"roomId"`
	Handle    string `json:"handle"`
}

type TalkInput struct {
	ID        *string `json:"id"`
	OwnerID   *string `json:"ownerId"`
	SpeakerID *string `json:"speakerId"`
	ConfaID   *string `json:"confaId"`
	Handle    *string `json:"handle"`
}

type Talks struct {
	Items    []*Talk `json:"items"`
	Limit    int     `json:"limit"`
	NextFrom string  `json:"nextFrom"`
}

type Token struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expiresIn"`
}
