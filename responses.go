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
