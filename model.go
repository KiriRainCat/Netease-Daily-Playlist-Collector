package main

type LoginStatus struct {
	Data struct {
		Code    int `json:"code"`
		Account struct {
			Id int `json:"id"`
		} `json:"account"`
		Profile struct {
			UserId   int    `json:"userId"`
			Nickname string `json:"nickname"`
		} `json:"profile"`
	} `json:"data"`
}

type QrKey struct {
	Data struct {
		Code   int    `json:"code"`
		UniKey string `json:"unikey"`
	} `json:"data"`
}

type QrImg struct {
	Code int `json:"code"`
	Data struct {
		QrUrl string `json:"qrurl"`
		QrImg string `json:"qrimg"`
	} `json:"data"`
}

type QrStatus struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Cookie  string `json:"cookie"`
}

type PlayListCreate struct {
	Code     int `json:"code"`
	Playlist struct {
		Id   int    `json:"id"`
		Name string `json:"name"`
	} `json:"playlist"`
}

type PlayListQuery struct {
	Code      int `json:"code"`
	Playlists []struct {
		Id   int    `json:"id"`
		Name string `json:"name"`
	} `json:"playlist"`
}

type DailySongs struct {
	Code int `json:"code"`
	Data struct {
		DailySongs []struct {
			Id   int    `json:"id"`
			Name string `json:"name"`
		} `json:"dailySongs"`
	} `json:"data"`
}
