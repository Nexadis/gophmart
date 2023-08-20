package user

import "encoding/json"

var (
	_ json.Marshaler   = &Balance{}
	_ json.Unmarshaler = &Balance{}
)

type Balance struct {
	Withdrawn int64
	Current   int64
}

type jsonBalance struct {
	Withdrawn float64
	Current   float64
}

func (b *Balance) MarshalJSON() ([]byte, error) {
	j := &jsonBalance{
		Withdrawn: float64(b.Withdrawn) / 100,
		Current:   float64(b.Current) / 100,
	}
	return json.Marshal(j)
}

func (b *Balance) UnmarshalJSON(data []byte) error {
	j := &jsonBalance{}
	err := json.Unmarshal(data, j)
	if err != nil {
		return err
	}
	b.Current = int64(j.Current) * 100
	b.Withdrawn = int64(j.Withdrawn) * 100
	return nil
}
