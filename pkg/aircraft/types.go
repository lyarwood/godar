package aircraft

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Aircraft represents the structure of an aircraft's data from the API
type Aircraft struct {
	ID           int           `json:"Id"`
	TSecs        int           `json:"TSecs"`
	Rcvr         int           `json:"Rcvr"`
	Icao         string        `json:"Icao,omitempty"`
	Bad          bool          `json:"Bad,omitempty"`
	Reg          string        `json:"Reg,omitempty"`
	Alt          int           `json:"Alt,omitempty"`
	GAlt         int           `json:"GAlt,omitempty"`
	InHg         float64       `json:"InHg,omitempty"`
	AltT         int           `json:"AltT,omitempty"`
	TAlt         int           `json:"TAlt,omitempty"`
	Call         string        `json:"Call,omitempty"`
	CallSus      bool          `json:"CallSus,omitempty"`
	Lat          float64       `json:"Lat,omitempty"`
	Long         float64       `json:"Long,omitempty"`
	PosTime      int64         `json:"PosTime,omitempty"`
	Mlat         bool          `json:"Mlat,omitempty"`
	PosStale     bool          `json:"PosStale,omitempty"`
	IsTisb       bool          `json:"IsTisb,omitempty"`
	Spd          float64       `json:"Spd,omitempty"`
	SpdTyp       int           `json:"SpdTyp,omitempty"`
	Vsi          int           `json:"Vsi,omitempty"`
	VsiT         int           `json:"VsiT,omitempty"`
	Trak         float64       `json:"Trak,omitempty"`
	TrkH         bool          `json:"TrkH,omitempty"`
	TTrk         float64       `json:"TTrk,omitempty"`
	Type         string        `json:"Type,omitempty"`
	Mdl          string        `json:"Mdl,omitempty"`
	Man          string        `json:"Man,omitempty"`
	CNum         string        `json:"CNum,omitempty"`
	From         string        `json:"From,omitempty"`
	To           string        `json:"To,omitempty"`
	Stops        []string      `json:"Stops,omitempty"`
	Op           string        `json:"Op,omitempty"`
	OpCode       string        `json:"OpCode,omitempty"`
	Sqk          SqkValue      `json:"Sqk,omitempty"`
	Help         bool          `json:"Help,omitempty"`
	Dst          float64       `json:"Dst,omitempty"`
	Brng         float64       `json:"Brng,omitempty"`
	WTC          WTCValue      `json:"WTC,omitempty"`
	Engines      string        `json:"Engines,omitempty"`
	EngType      EngTypeValue  `json:"EngType,omitempty"`
	EngMount     EngMountValue `json:"EngMount,omitempty"`
	Species      SpeciesValue  `json:"Species,omitempty"`
	Mil          bool          `json:"Mil,omitempty"`
	Cou          string        `json:"Cou,omitempty"`
	HasPic       bool          `json:"HasPic,omitempty"`
	PicX         int           `json:"PicX,omitempty"`
	PicY         int           `json:"PicY,omitempty"`
	FlightsCount int           `json:"FlightsCount,omitempty"`
	CMsgs        int           `json:"CMsgs,omitempty"`
	Gnd          bool          `json:"Gnd,omitempty"`
	Tag          string        `json:"Tag,omitempty"`
	Interested   bool          `json:"Interested,omitempty"`
	TT           string        `json:"TT,omitempty"`
	Trt          int           `json:"Trt,omitempty"`
	Year         string        `json:"Year,omitempty"`
	Sat          bool          `json:"Sat,omitempty"`
	Cos          []float64     `json:"Cos,omitempty"`
	Cot          []float64     `json:"Cot,omitempty"`
	ResetTrail   bool          `json:"ResetTrail,omitempty"`
	HasSig       bool          `json:"HasSig,omitempty"`
	Sig          float64       `json:"Sig,omitempty"`
}

// Feed represents the structure of a feed object
type Feed struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// AircraftList represents the top-level structure of the JSON response from the server
type AircraftList struct {
	LastDv        LastDvValue `json:"lastDv"`
	TotalAc       int         `json:"totalAc"`
	Src           int         `json:"src"`
	ShowSil       bool        `json:"showSil,omitempty"`
	ShowFlg       bool        `json:"showFlg,omitempty"`
	ShowPic       bool        `json:"showPic,omitempty"`
	FlgW          int         `json:"flgW,omitempty"`
	FlgH          int         `json:"flgH,omitempty"`
	ShtTrlSec     int         `json:"shtTrlSec,omitempty"`
	Stm           int64       `json:"stm"`
	Aircraft      []Aircraft  `json:"acList"`
	Feeds         []Feed      `json:"feeds,omitempty"`
	SrcFeed       int         `json:"srcFeed,omitempty"`
	ConfigChanged bool        `json:"configChanged,omitempty"`
}

// LastDvValue handles both string and number formats for lastDv field
type LastDvValue int64

// UnmarshalJSON implements custom unmarshaling for LastDvValue
func (ldv *LastDvValue) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var strVal string
	if err := json.Unmarshal(data, &strVal); err == nil {
		// Convert string to int64
		intVal, err := strconv.ParseInt(strVal, 10, 64)
		if err != nil {
			return err
		}
		*ldv = LastDvValue(intVal)
		return nil
	}

	// If string fails, try as number
	var intVal int64
	if err := json.Unmarshal(data, &intVal); err != nil {
		return err
	}
	*ldv = LastDvValue(intVal)
	return nil
}

// Int64 returns the LastDvValue as int64
func (ldv LastDvValue) Int64() int64 {
	return int64(ldv)
}

// SqkValue handles both string and number formats for the squawk code
type SqkValue int

// UnmarshalJSON implements custom unmarshaling for SqkValue
func (s *SqkValue) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var strVal string
	if err := json.Unmarshal(data, &strVal); err == nil {
		if strVal == "" {
			*s = 0
			return nil
		}
		intVal, err := strconv.Atoi(strVal)
		if err != nil {
			return err
		}
		*s = SqkValue(intVal)
		return nil
	}
	// If string fails, try as number
	var intVal int
	if err := json.Unmarshal(data, &intVal); err != nil {
		return err
	}
	*s = SqkValue(intVal)
	return nil
}

// WTCValue handles both string and number formats for the WTC field
type WTCValue string

// UnmarshalJSON implements custom unmarshaling for WTCValue
func (w *WTCValue) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var strVal string
	if err := json.Unmarshal(data, &strVal); err == nil {
		*w = WTCValue(strVal)
		return nil
	}
	// If string fails, try as number
	var numVal int
	if err := json.Unmarshal(data, &numVal); err == nil {
		*w = WTCValue(strconv.Itoa(numVal))
		return nil
	}
	return fmt.Errorf("WTCValue: could not unmarshal %s", string(data))
}

// SpeciesValue handles both string and number formats for the Species field
type SpeciesValue string

// UnmarshalJSON implements custom unmarshaling for SpeciesValue
func (s *SpeciesValue) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var strVal string
	if err := json.Unmarshal(data, &strVal); err == nil {
		*s = SpeciesValue(strVal)
		return nil
	}
	// If string fails, try as number
	var numVal int
	if err := json.Unmarshal(data, &numVal); err == nil {
		*s = SpeciesValue(strconv.Itoa(numVal))
		return nil
	}
	return fmt.Errorf("SpeciesValue: could not unmarshal %s", string(data))
}

// EngTypeValue handles both string and number formats for the EngType field
type EngTypeValue string

// UnmarshalJSON implements custom unmarshaling for EngTypeValue
func (e *EngTypeValue) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var strVal string
	if err := json.Unmarshal(data, &strVal); err == nil {
		*e = EngTypeValue(strVal)
		return nil
	}
	// If string fails, try as number
	var numVal int
	if err := json.Unmarshal(data, &numVal); err == nil {
		*e = EngTypeValue(strconv.Itoa(numVal))
		return nil
	}
	return fmt.Errorf("EngTypeValue: could not unmarshal %s", string(data))
}

// EngMountValue handles both string and number formats for the EngMount field
type EngMountValue string

// UnmarshalJSON implements custom unmarshaling for EngMountValue
func (e *EngMountValue) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var strVal string
	if err := json.Unmarshal(data, &strVal); err == nil {
		*e = EngMountValue(strVal)
		return nil
	}
	// If string fails, try as number
	var numVal int
	if err := json.Unmarshal(data, &numVal); err == nil {
		*e = EngMountValue(strconv.Itoa(numVal))
		return nil
	}
	return fmt.Errorf("EngMountValue: could not unmarshal %s", string(data))
}
