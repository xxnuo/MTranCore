package main

import "encoding/json"

type PoweronRequest struct {
	Path                 string   `json:"path,omitempty"`
	ModelPath            string   `json:"model_path,omitempty"`
	LexicalShortlistPath string   `json:"lexical_shortlist_path,omitempty"`
	VocabularyPath       string   `json:"vocabulary_path,omitempty"`
	VocabularyPaths      []string `json:"vocabulary_paths,omitempty"`
}

type ExitRequest struct {
        Time  int  `json:"time"`
        Force bool `json:"force"`
}

type TransRequest struct {
        Text string `json:"text"`
        HTML bool   `json:"html,omitempty"`
}
type TransResponse struct {
	TranslatedText string `json:"translated_text"`
}

type HealthResponse struct {
	Health bool `json:"health"`
}

type WSMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

type WSResponse struct {
	Type string      `json:"type"`
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}
