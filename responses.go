package tracker

type ErrorResponse struct {
	FailureReason string `bencode:"failure reason"`
}

type AnnounceResponse struct {
	Interval    int    `bencode:"interval"`
	MinInterval int    `bencode:"min interval"`
	Complete    int    `bencode:"complete"`
	Incomplete  int    `bencode:"incomplete"`
	Peers       string `bencode:"peers"`
}

type ScrapeResponse struct {
	Files map[string]ScrapeTorrent `bencode:"files"`
}

type ScrapeTorrent struct {
	Complete   int `bencode:"complete"`
	Incomplete int `bencode:"incomplete"`
	Downloaded int `bencode:"downloaded"`
}
