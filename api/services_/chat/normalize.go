// services_/chat/normalize.go
package chat

import "log"

func normalizeMessageEncryption(msg *Message) {
	if msg.IsEncrypted {
		msg.Content = ""
		for _, att := range msg.Attachments {
			att.IsEncrypted = true
			log.Printf("Set attachment IsEncrypted: true, att.Envelope: %+v, msg.Envelope: %+v", att.Envelope, msg.Envelope)
		}
	} else {
		for _, att := range msg.Attachments {
			att.IsEncrypted = false
			log.Printf("Set attachment IsEncrypted: false, att.Envelope: %+v, msg.Envelope: %+v", att.Envelope, msg.Envelope)
		}
	}
}
