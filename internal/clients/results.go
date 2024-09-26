package clients

type resourceIdentifier struct {
	RID   string `json:"rid"`
	RType string `json:"rtype"`
}

type position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type EntertainmentConfiguration struct {
	Type     string `json:"type"`
	ID       string `json:"id"`
	IDV1     string `json:"id_v1"`
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Name              string             `json:"name"`
	ConfigurationType string             `json:"configuration_type"`
	Status            string             `json:"status"`
	ActiveStreamer    resourceIdentifier `json:"active_streamer"`
	StreamProxy       struct {
		Mode string             `json:"mode"`
		Node resourceIdentifier `json:"node"`
	} `json:"stream_proxy"`
	Channels []struct {
		ChannelID uint8    `json:"channel_id"`
		Position  position `json:"position"`
		Members   [](struct {
			Service resourceIdentifier `json:"service"`
			Index   int                `json:"index"`
		}) `json:"members"`
	} `json:"channels"`
	Locations struct {
		ServiceLocations []struct {
			Service            resourceIdentifier `json:"service"`
			Position           position           `json:"position"`
			Positions          []position         `json:"positions"`
			EqualizationFactor float64            `json:"equalization_factor"`
		} `json:"service_locations"`
	} `json:"locations"`
	LightServices []resourceIdentifier `json:"light_services"`
}

type EntertainmentConfigurationResult struct {
	Errors [](struct {
		Description string `json:"description"`
	}) `json:"errors"`
	Data []EntertainmentConfiguration `json:"data"`
}
