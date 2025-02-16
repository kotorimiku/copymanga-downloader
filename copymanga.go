package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/tidwall/gjson"
)

func Register(user *User) error {
	username := GenerateUsername(8)
	password := GeneratePassword(16)

	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)

	url := fmt.Sprintf("https://%s/api/v3/register", ConfigInstance.UrlBase)
	res, err := client.PostForm(url, data)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		if user == nil {
			user = &User{}
		}
		user.Username = username
		user.Password = password
		return nil
	}
	return fmt.Errorf("注册失败")
}

func Login(user *User) error {
	salt := "582496"

	password := Password(user.Password, salt)

	data := url.Values{}
	data.Set("username", user.Username)
	data.Set("password", password)
	data.Set("salt", salt)

	url := fmt.Sprintf("https://%s/api/v3/login", ConfigInstance.UrlBase)

	res, err := client.PostForm(url, data)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	token := gjson.GetBytes(body, "results.token")
	user.Token = token.String()

	return nil
}

func Password(password string, salt string) string {
	data := fmt.Sprintf("%s-%s", password, salt)

	encoded := base64.StdEncoding.EncodeToString([]byte(data))

	return encoded
}
