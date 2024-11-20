package ws

type OjoWebsocketBroadcast struct {
	client *OjoWebsocketClient
}

func (b *OjoWebsocketBroadcast) Emit(eventName string, payload ...(any)) {
	for _, client := range b.client.ojoWS.clients {
		if b.client.Id != client.Id {
			client.Emit(eventName, payload...)
		}
	}
}
