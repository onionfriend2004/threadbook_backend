package dto

type ConnectAndSubscribeTokens struct {
	ConnectToken  string            `json:"connect_token"`
	ChannelTokens map[string]string `json:"channel_tokens"` // ключ = канал, значение = JWT токен
}
