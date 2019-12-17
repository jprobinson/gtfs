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

		Synonyms []string
	}

	Synonym struct {
		Value    string   `json:"value"`
		Synonyms []string `json:"synonyms"`
	}
)
