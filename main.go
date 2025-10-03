package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"text/template"
	"time"
)

type TwitchAuth struct {
	Token     string `json:"access_token"`
	ExpiresIn int64  `json:"expires_in"`
	ExpiresAt time.Time
}

type TwitchStream struct {
	UserName string `json:"user_name"`
	Type     string `json:"type"`
}

type TwitchStreamsResponse struct {
	Data []TwitchStream `json:"data"`
}

func main() {

	clientID, clientIDExists := os.LookupEnv("TWITCH_CLIENT_ID")
	if !clientIDExists {
		fmt.Println("Missing TWITCH_CLIENT_ID environment variable")
	}
	clientSecret, clientSecretExists := os.LookupEnv("TWITCH_CLIENT_SECRET")
	if !clientSecretExists {
		fmt.Println("Missing TWITCH_CLIENT_SECRET environment variable")
	}
	twitchChannel, twitchChannelExists := os.LookupEnv("TWITCH_CHANNEL")
	if !twitchChannelExists {
		fmt.Println("Missing TWITCH_CHANNEL environment variable")
	}
	port, portExists := os.LookupEnv("PORT")
	if !portExists {
		port = "8080"
	}

	if !clientIDExists || !clientSecretExists || !twitchChannelExists {
		os.Exit(1)
	}

	auth := &TwitchAuth{
		ExpiresAt: time.Now(),
	}

	getToken := func() (string, error) {
		if auth != nil && time.Now().Before(auth.ExpiresAt) {
			fmt.Println("using cached token")
			return auth.Token, nil
		}
		resp, err := http.Post("https://id.twitch.tv/oauth2/token?client_id="+clientID+"&client_secret="+clientSecret+"&grant_type=client_credentials", "application/json", nil)
		if err != nil {
			return "", fmt.Errorf("failed to authenticate with twitch token: %v\n", err)
		}

		respBody := make([]byte, resp.ContentLength)
		_, err = resp.Body.Read(respBody)
		if err != nil {
			return "", fmt.Errorf("failed to read response body: %v\n", err)
		}

		err = json.Unmarshal(respBody, auth)
		if err != nil {
			return "", fmt.Errorf("failed to unmarshal twitch response: %v\n", err)
		}

		now := time.Now()
		auth.ExpiresAt = now.Add(time.Duration(auth.ExpiresIn) * time.Second).Add(-2 * time.Minute)

		return auth.Token, nil
	}

	server := http.NewServeMux()

	server.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		token, err := getToken()
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
		}

		req, err := http.NewRequest("GET", "https://api.twitch.tv/helix/streams?user_login="+twitchChannel, nil)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
		}
		req.Header.Add("Authorization", "Bearer "+token)
		req.Header.Add("Client-Id", clientID)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
		}

		respBody := make([]byte, resp.ContentLength)
		resp.Body.Read(respBody)

		streams := &TwitchStreamsResponse{}
		json.Unmarshal(respBody, streams)

		isLive := false
		for _, stream := range streams.Data {
			isLive = stream.Type == "live"
		}

		tmpl, err := template.ParseFiles("index.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error loading template: " + err.Error()))
			return
		}

		places := []string{
			"a live tá debaixo da sua cama",
			"a live tá no bolso de trás da sua calça",
			"a live tá dentro da gaveta de meias da sua mãe",
			"a live tá dentro do diário secreto do seu gato",
			"a live tá atrás do ímã de geladeira de lembrança de Ilha Bela",
			"a live tá dentro do balde de pipoca no cinema",
			"a live tá embaixo da almofada do sofá",
			"a live tá escondida na gaveta de bagunça que tem mais bagunça do que gaveta",
			"a live tá escondida shhh 🤫",
			"a live tá flutuando na garrafa de água na geladeira do seu vizinho",
			"pergunta pro omegamain",
			"a live tá presa na coleira do seu cachorro",
			"a live tá dentro da lata de leite ninho",
			"a live tá entre as almofadas do sofá do seu tio-avô",
			"ele não pode agora, ele tá sentado na beira do universo, esperando um ônibus",
			"a live tá atrás da moldura daquela foto de família",
			`<a href="https://www.youtube.com/watch?v=--9kqhzQ-8Q">H.Y.C.Y.BH?</a>`,
		}

		// pick one place at random
		place := places[time.Now().UnixNano()%int64(len(places))]

		data := struct {
			Place  string
			IsLive bool
		}{
			IsLive: isLive,
			Place:  place,
		}

		w.Header().Set("Content-Type", "text/html")
		err = tmpl.Execute(w, data)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error rendering template: " + err.Error()))
			return
		}
	})

	http.ListenAndServe(":"+port, server)
}
