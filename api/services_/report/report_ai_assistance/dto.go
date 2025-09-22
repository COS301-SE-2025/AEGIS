package report_ai_assistance

// ReportSectionResponse is the JSON payload sent to the frontend
type ReportSectionResponse struct {
	ID      string `json:"id"`   // Mongo ObjectID (stringified) if you’re still storing sections in Mongo
	UUID    string `json:"uuid"` // Postgres UUID from report_ai_assistance.ReportSection
	Title   string `json:"title"`
	Content string `json:"content"`
}
