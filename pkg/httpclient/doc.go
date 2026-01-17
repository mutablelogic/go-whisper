// Package httpclient provides a typed Go client for consuming the go-whisper
// REST API.
//
// Create a client with:
//
//	client, err := httpclient.New("http://localhost:8080/api/whisper")
//	if err != nil {
//	   panic(err)
//	}
//
// Then use the client to manage models and transcribe audio:
//
//	// List all models
//	models, err := client.ListModels(ctx)
//
//	// Get a specific model
//	model, err := client.GetModel(ctx, "tiny")
//
//	// Download a model with progress reporting
//	model, err := client.DownloadModel(ctx, "tiny", func(cur, total uint64) {
//	    fmt.Printf("Progress: %d/%d bytes\n", cur, total)
//	})
//
//	// Delete a model
//	err := client.DeleteModel(ctx, "tiny")
//
//	// Transcribe audio from a file
//	file, _ := os.Open("audio.mp3")
//	result, err := client.Transcribe(ctx, "tiny", file)
//
//	// Transcribe with language specification
//	result, err := client.Transcribe(ctx, "tiny", file,
//	    httpclient.WithLanguage("en"))
//
//	// Translate audio to English
//	result, err := client.Translate(ctx, "tiny", file)
//
//	// Get transcription in specific format
//	result, err := client.Transcribe(ctx, "tiny", file,
//	    httpclient.WithFormat(httpclient.FormatVTT))
package httpclient
