package main

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/skip2/go-qrcode"
)

var api *resty.Client

func init() {
	api = resty.New()
	api.BaseURL = "https://netease.kiriraincat.eu.org"

	data, err := os.ReadFile("cookie.txt")
	if err == nil {
		api.SetQueryParam("cookie", url.QueryEscape(string(data)))
	}

	api.OnError(func(req *resty.Request, err error) {
		fmt.Println(err.Error())
		fmt.Println("网络错误或远程端点API离线 (请联系开发者: 柒夜雨猫)")
	})
}

func main() {
	//* -------------------------------- 检查登录状态 -------------------------------- *//
	fmt.Println("检查登录状态中...")

	var user LoginStatus
	api.R().
		SetResult(&user).
		SetQueryParam("timestamp", fmt.Sprint(time.Now().Unix())).
		Get("/login/status")

	if user.Data.Account.Id == 0 {
		fmt.Println("登录状态不存在或过期，请扫码登录喵~")
		qrLogin()
	}

	//* ------------------------------- 添加今日推荐到歌单 ------------------------------ *//
	fmt.Println("已登录，获取今日推荐曲目并添加到歌单中...")

	// 判断今日是否已经添加歌单
	var playlists PlayListQuery
	api.R().
		SetResult(&playlists).
		SetQueryParam("timestamp", fmt.Sprint(time.Now().Unix())).
		SetQueryParam("uid", strconv.Itoa(user.Data.Profile.UserId)).
		Get("/user/playlist")

	playlistName := fmt.Sprint("每日推荐 ", time.Now().Format("2006-01-02"))
	for _, playlist := range playlists.Playlists {
		if playlist.Name == playlistName {
			fmt.Printf("哎呀，%s - 今天已经添加过推荐曲目啦，请不要重复添加喵~\n", user.Data.Profile.Nickname)
			fmt.Println("\n将在 5 秒后自动退出~")
			for i := 5; i >= 0; i-- {
				fmt.Printf("倒计时 %d 秒...\n", i)
				time.Sleep(time.Second * 1)
			}
			return
		}
	}

	// 创建歌单
	var playlist PlayListCreate
	api.R().
		SetResult(&playlist).
		SetQueryParam("timestamp", fmt.Sprint(time.Now().Unix())).
		SetQueryParam("name", playlistName).
		Get("/playlist/create")

	// 获取今日推荐歌曲
	var songs DailySongs
	api.R().
		SetResult(&songs).
		SetQueryParam("timestamp", fmt.Sprint(time.Now().Unix())).
		Get("/recommend/songs")

	// 添加今日推荐歌曲至歌单
	fmt.Printf("添加今日推荐歌曲至 歌单[%s] 中...\n", playlist.Playlist.Name)

	for _, song := range songs.Data.DailySongs {
		api.R().
			SetQueryParam("timestamp", fmt.Sprint(time.Now().Unix())).
			SetQueryParam("op", "add").
			SetQueryParam("pid", strconv.Itoa(playlist.Playlist.Id)).
			SetQueryParam("tracks", strconv.Itoa(song.Id)).
			Get("/playlist/tracks")
		fmt.Println("成功添加曲目:", song.Name)
	}

	fmt.Printf("%s - 今天推荐歌单添加成功啦，快去听歌叭~\n", user.Data.Profile.Nickname)
	fmt.Println("\n将在 5 秒后自动退出~")
	for i := 5; i >= 0; i-- {
		fmt.Printf("倒计时 %d 秒...\n", i)
		time.Sleep(time.Second * 1)
	}
}

func genQr() string {
	var key QrKey
	api.R().
		SetResult(&key).
		SetQueryParam("timestamp", fmt.Sprint(time.Now().Unix())).
		Get("/login/qr/key")

	var img QrImg
	api.R().
		SetResult(&img).
		SetQueryParam("timestamp", fmt.Sprint(time.Now().Unix())).
		SetQueryParam("key", key.Data.UniKey).
		SetQueryParam("noCookie", "true").
		Get("/login/qr/create")

	qr, _ := qrcode.New(img.Data.QrUrl, qrcode.High)
	fmt.Println(qr.ToSmallString(true))

	return key.Data.UniKey
}

func qrLogin() {
	key := genQr()

	var status QrStatus
	for i := 0; i < 1; {
		time.Sleep(time.Second * 4)
		api.R().
			SetResult(&status).
			SetQueryParam("timestamp", fmt.Sprint(time.Now().Unix())).
			SetQueryParam("key", key).
			Get("/login/qr/check")

		if status.Code != 801 && status.Code != 802 {
			fmt.Println(status.Message)

			if status.Code == 800 {
				fmt.Print(" → 请重新扫码登录呐 = =")
				key = genQr()
			} else {
				for _, cookie := range strings.Split(status.Cookie, ";") {
					if strings.Contains(cookie, "MUSIC_U") {
						api.SetQueryParam("cookie", status.Cookie)
						os.WriteFile("cookie.txt", []byte(status.Cookie), os.ModeDevice)
						break
					}
				}
				break
			}
		}
	}
}
