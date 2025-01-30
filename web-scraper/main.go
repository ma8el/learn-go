package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/charmbracelet/huh"
)

type LocalAIRequest struct {
	Model    string                 `json:"model"`
	Options  map[string]interface{} `json:"options"`
	Messages []struct {
		Role    string `json:"role"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"messages"`
	Temperature float32 `json:"temperature"`
}

type LocalAIResponse struct {
	Created int64  `json:"created"`
	Object  string `json:"object"`
	ID      string `json:"id"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

var localAIUrl = "http://localhost:11434/v1/chat/completions"

//var context = `I am interested in python, typescript, vuejs and go.
//Give me the most interesting topics and the most interesting posts.
//I am not interested in improving the html code.
//The html does not interest me.`
//
//var model = "marco-o1"
//var temperature float32 = 0.5

var (
	model       string
	context     string
	temperature string
)

func main() {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Choose a model").
				Options(
					huh.NewOption("marco-o1", "marco-o1"),
					huh.NewOption("deepseek-r1", "deepseek-r1"),
				).Value(&model),
			huh.NewInput().Title("Context").Placeholder("Enter context").Value(&context),
			huh.NewInput().Title("Temperature").Placeholder("Enter temperature").Value(&temperature),
		),
	)

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}

	temp, err := strconv.ParseFloat(temperature, 32)
	if err != nil {
		log.Fatal(err)
	}

	response := getAICuratedPosts(context, model, float32(temp))
	fmt.Println(response)
}

func getAICuratedPosts(context string, model string, temperature float32) string {
	doc := getPageDocument("https://news.ycombinator.com/")
	posts := getHackerNewsPosts(doc)
	response := getLocalAIResponse(posts, context, model, temperature)
	return response
}

func getPageDocument(url string) *goquery.Document {
	response, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	return doc
}

func getPageLinks(doc *goquery.Document) []string {
	links := []string{}
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		link, _ := s.Attr("href")
		if len(link) >= 5 && link[:5] == "https" {
			links = append(links, link)
		}
	})
	return links
}

func getPageHTML(doc *goquery.Document) string {
	html, err := doc.Html()
	if err != nil {
		log.Fatal(err)
	}
	return html
}

func getHackerNewsPosts(doc *goquery.Document) string {
	var posts string

	doc.Find("tr.athing").Each(func(i int, s *goquery.Selection) {
		title := s.Find("span.titleline").Text()
		plainLink, _ := s.Find("td.title").Find("span.titleline").Find("a").Attr("href")
		rank := s.Find("span.rank").Text()
		posts += rank + ": " + title + " " + plainLink + "\n"
	})
	return posts
}

func getLocalAIResponse(message string, context string, model string, temperature float32) string {
	if model == "" {
		model = "gpt-4"
	}
	if temperature == 0 {
		temperature = 0.5
	}
	if context == "" {
		context = "You are a helpful assistant."
	}

	requestBody := LocalAIRequest{
		Model: model,
		Messages: []struct {
			Role    string `json:"role"`
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		}{
			{
				Role: "user",
				Content: []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				}{
					{
						Type: "text",
						Text: context,
					},
					{
						Type: "text",
						Text: message,
					},
				},
			},
		},
		Temperature: temperature,
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		log.Fatal(err)
		return "Error marshalling request body, " + err.Error()
	}
	response, err := http.Post(localAIUrl, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Fatal(err)
		return "Error sending request, " + err.Error()
	}
	defer response.Body.Close()

	var localAIResponse LocalAIResponse
	err = json.NewDecoder(response.Body).Decode(&localAIResponse)
	if err != nil {
		log.Fatal(err)
		return "Error decoding response, " + err.Error()
	}

	if len(localAIResponse.Choices) == 0 {
		return "No response received from AI"
	}

	return localAIResponse.Choices[0].Message.Content
}
