package fc2

type WhatLinkInfo struct {
	Error       string `json:"error"`
	Type        string `json:"type"`
	FileType    string `json:"file_type"`
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	Count       int    `json:"count"`
	Screenshots []struct {
		Time       int    `json:"time"`
		Screenshot string `json:"screenshot"`
	} `json:"screenshots"`
}
