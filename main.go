package main

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/skip2/go-qrcode"
)

var api *resty.Client

var user LoginStatus

func init() {
	api = resty.New()
	api.BaseURL = "https://netease.kiriraincat.eu.org"

	cwd, _ := os.Executable()
	data, err := os.ReadFile(path.Dir(cwd) + "/cookie.txt")
	if err == nil {
		api.SetQueryParam("cookie", string(data))
	}

	api.OnError(func(req *resty.Request, err error) {
		fmt.Println(err.Error())
		fmt.Println("网络错误或远程端点API离线 (请联系开发者: 柒夜雨猫)")
	})
}

func main() {
	//* -------------------------------- 检查登录状态 -------------------------------- *//
	fmt.Println("检查登录状态中...")

	api.R().
		SetResult(&user).
		SetQueryParam("timestamp", fmt.Sprint(time.Now().Unix())).
		Get("/login/status")

	if user.Data.Account.Id == 0 {
		fmt.Println("登录状态不存在或过期，请选择登录方式喵~ (输入 1 到 4)")
		fmt.Println("1. 扫码登录")
		fmt.Println("2. 手机验证码登录")
		fmt.Println("3. 手机密码登录")
		fmt.Println("4. 邮箱密码登录")

		func() {
			for i := 0; i < 1; {
				var choice string
				fmt.Scanln(&choice)

				// 校验输入
				switch choice {
				case "1":
					qrLogin()
					return
				case "2":
					phoneLogin()
					return
				case "3":
					phonePwdLogin()
					return
				case "4":
					emailPwdLogin()
					return
				default:
					fmt.Println("输入错误，请重新输入喵~")
				}
			}
		}()
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

func storeCookie(raw string) {
	for _, cookie := range strings.Split(raw, ";") {
		if strings.Contains(cookie, "MUSIC_U") {
			api.SetQueryParam("cookie", cookie)
			cwd, _ := os.Executable()
			os.WriteFile(path.Dir(cwd)+"/cookie.txt", []byte(cookie), 0777)
			break
		}
	}

	api.R().
		SetResult(&user).
		SetQueryParam("timestamp", fmt.Sprint(time.Now().Unix())).
		Get("/login/status")

	if user.Data.Account.Id == 0 {
		fmt.Println("登录异常")
		fmt.Println("\n将在 5 秒后自动退出...")
		os.Exit(0)
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
				storeCookie(status.Cookie)
				break
			}
		}
	}
}

func phoneLogin() {
	var phone string
	fmt.Print("请输入手机号码: ")
	fmt.Scanln(&phone)

	// 调用网易云接口发送验证码
	api.R().
		SetQueryParam("timestamp", fmt.Sprint(time.Now().Unix())).
		SetQueryParam("phone", phone).
		Get("/captcha/sent")

	fmt.Print("请输入验证码: ")
	var code string
	fmt.Scanln(&code)

	var res PhoneLogin
	api.R().
		SetResult(&res).
		SetQueryParam("timestamp", fmt.Sprint(time.Now().Unix())).
		SetQueryParam("phone", phone).
		SetQueryParam("captcha", code).
		Get("/login/cellphone")

	storeCookie(res.Cookie)
}

func phonePwdLogin() {
	var phone string
	fmt.Print("请输入手机号码: ")
	fmt.Scanln(&phone)

	var pwd string
	fmt.Print("请输入密码: ")
	fmt.Scanln(&pwd)

	var res PhoneLogin
	api.R().
		SetResult(&res).
		SetQueryParam("timestamp", fmt.Sprint(time.Now().Unix())).
		SetQueryParam("phone", phone).
		SetQueryParam("password", pwd).
		Get("/login/cellphone")

	storeCookie(res.Cookie)
}

func emailPwdLogin() {
	var email string
	fmt.Print("请输入网易邮箱: ")
	fmt.Scanln(&email)

	var pwd string
	fmt.Print("请输入密码: ")
	fmt.Scanln(&pwd)

	var res PhoneLogin
	api.R().
		SetResult(&res).
		SetQueryParam("timestamp", fmt.Sprint(time.Now().Unix())).
		SetQueryParam("email", email).
		SetQueryParam("password", pwd).
		Get("/login")

	storeCookie(res.Cookie)
}
