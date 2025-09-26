package chat

import "log"

func NormalizeMessageEncryption(msg *Message) {
	if msg.IsEncrypted {
		if msg.MessageType == "text" {
			msg.Content = "" // Clear content only for text messages
		}
		for _, att := range msg.Attachments {
			if msg.IsEncrypted || att.Envelope != nil || msg.Envelope != nil {
				att.IsEncrypted = true
				log.Printf("Set attachment IsEncrypted: true, att.Envelope: %+v, msg.Envelope: %+v", att.Envelope, msg.Envelope)
			}
		}
	}
}
