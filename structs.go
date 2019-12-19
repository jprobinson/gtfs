package gtfs

type (
	Route struct {
		Name string

		Northbound string
		Southbound string

		Stops []Stop
	}

	Stop struct {
		ID string

		MTAName      string
		DisplayName  string
		PhoneticName string

		Synonyms  []string
		Transfers []Transfer `json:",omitempty"`
	}

	Transfer struct {
		StopID string
		Route  string
	}

	Synonym struct {
		Value    string   `json:"value"`
		Synonyms []string `json:"synonyms"`
	}
)
