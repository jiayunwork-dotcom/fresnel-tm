package recipe

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io"
	"math"
	"os"

	"fresnel-tm/internal/model"
	"fresnel-tm/internal/optics"
)

const magic = "FRTM"

const version uint16 = 1

type Record struct {
	Incident     model.Material `json:"incident"`
	Layers       model.Layers   `json:"layers"`
	Substrate    model.Material `json:"substrate"`
	WavelengthNm float64        `json:"wavelength_nm"`
	AngleDeg     float64        `json:"angle_deg"`
	Polarization string         `json:"polarization"`
	Reflection   float64        `json:"reflection"`
	Transmission float64        `json:"transmission"`
	Absorption   float64        `json:"absorption"`
}

func (r Record) Stack() model.Stack {
	return model.Stack{Incident: r.Incident, Layers: r.Layers, Substrate: r.Substrate}
}

func (r Record) Incidence() model.Incidence {
	return model.Incidence{WavelengthNm: r.WavelengthNm, AngleDeg: r.AngleDeg}
}

func (r Record) Validate() error {
	if err := r.Stack().Validate("配方膜系"); err != nil {
		return err
	}
	if err := r.Incidence().Validate("配方入射"); err != nil {
		return err
	}
	if _, err := model.ParsePolarization(r.Polarization); err != nil {
		return err
	}
	for _, v := range []float64{r.Reflection, r.Transmission, r.Absorption} {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return fmt.Errorf("recipe: optical result is not finite")
		}
		if v < -1e-9 || v > 1+1e-9 {
			return fmt.Errorf("recipe: optical result out of [0,1]")
		}
	}
	return nil
}

func Seal(s model.Stack, inc model.Incidence, pol model.Polarization) (Record, error) {
	if err := s.Validate("膜系"); err != nil {
		return Record{}, err
	}
	if err := inc.Validate("入射"); err != nil {
		return Record{}, err
	}
	res, err := optics.Solve(s, inc, pol)
	if err != nil {
		return Record{}, err
	}
	copied := s.WithLayers(s.Layers)
	rec := Record{
		Incident:     s.Incident,
		Layers:       copied.Layers,
		Substrate:    s.Substrate,
		WavelengthNm: inc.WavelengthNm,
		AngleDeg:     inc.AngleDeg,
		Polarization: pol.String(),
		Reflection:   res.Reflection,
		Transmission: res.Transmission,
		Absorption:   res.Absorption,
	}
	if err := rec.Validate(); err != nil {
		return Record{}, err
	}
	return rec, nil
}

func Verify(rec Record) error {
	if err := rec.Validate(); err != nil {
		return err
	}
	pol, err := model.ParsePolarization(rec.Polarization)
	if err != nil {
		return err
	}
	res, err := optics.Solve(rec.Stack(), rec.Incidence(), pol)
	if err != nil {
		return err
	}
	if math.Abs(res.Reflection-rec.Reflection) > 1e-9 {
		return fmt.Errorf("recipe: stored R %g != solved R %g", rec.Reflection, res.Reflection)
	}
	if math.Abs(res.Transmission-rec.Transmission) > 1e-9 {
		return fmt.Errorf("recipe: stored T %g != solved T %g", rec.Transmission, res.Transmission)
	}
	if math.Abs(res.Absorption-rec.Absorption) > 1e-9 {
		return fmt.Errorf("recipe: stored A %g != solved A %g", rec.Absorption, res.Absorption)
	}
	return nil
}

func encodePayload(rec Record) ([]byte, error) {
	if err := rec.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(rec)
}

func decodePayload(p []byte) (Record, error) {
	var rec Record
	if err := json.Unmarshal(p, &rec); err != nil {
		return Record{}, fmt.Errorf("recipe: payload is not a coating record: %w", err)
	}
	if err := rec.Validate(); err != nil {
		return Record{}, err
	}
	return rec, nil
}

func writeHeader(f *os.File) error {
	if _, err := f.Write([]byte(magic)); err != nil {
		return err
	}
	var ver [2]byte
	binary.LittleEndian.PutUint16(ver[:], version)
	_, err := f.Write(ver[:])
	return err
}

func readHeader(r io.Reader) error {
	var mag [4]byte
	if _, err := io.ReadFull(r, mag[:]); err != nil {
		return fmt.Errorf("recipe: missing FRTM header: %w", err)
	}
	if string(mag[:]) != magic {
		return fmt.Errorf("recipe: bad magic %q", mag)
	}
	var ver [2]byte
	if _, err := io.ReadFull(r, ver[:]); err != nil {
		return fmt.Errorf("recipe: truncated version: %w", err)
	}
	got := binary.LittleEndian.Uint16(ver[:])
	if got != version {
		return fmt.Errorf("recipe: unsupported version %d", got)
	}
	return nil
}

func appendRecord(f *os.File, rec Record) error {
	payload, err := encodePayload(rec)
	if err != nil {
		return err
	}
	sum := crc32.ChecksumIEEE(payload)
	var hdr [8]byte
	binary.LittleEndian.PutUint32(hdr[0:4], uint32(len(payload)))
	binary.LittleEndian.PutUint32(hdr[4:8], sum)
	if _, err := f.Write(hdr[:]); err != nil {
		return err
	}
	_, err = f.Write(payload)
	return err
}

func Create(path string, rec Record) error {
	if err := rec.Validate(); err != nil {
		return err
	}
	if err := Verify(rec); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	if err := writeHeader(f); err != nil {
		return err
	}
	if err := appendRecord(f, rec); err != nil {
		return err
	}
	return f.Sync()
}

func Commit(path string, rec Record) error {
	if err := rec.Validate(); err != nil {
		return err
	}
	if err := Verify(rec); err != nil {
		return err
	}
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return Create(path, rec)
		}
		return err
	}
	f, err := os.OpenFile(path, os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	if err := readHeader(f); err != nil {
		return err
	}
	if _, err := f.Seek(0, io.SeekEnd); err != nil {
		return err
	}
	if err := appendRecord(f, rec); err != nil {
		return err
	}
	return f.Sync()
}

func Replay(path string) ([]Record, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	if err := readHeader(f); err != nil {
		return nil, err
	}
	var out []Record
	for {
		var hdr [8]byte
		n, err := io.ReadFull(f, hdr[:])
		if err == io.EOF || (err == io.ErrUnexpectedEOF && n == 0) {
			break
		}
		if err != nil {
			break
		}
		ln := binary.LittleEndian.Uint32(hdr[0:4])
		want := binary.LittleEndian.Uint32(hdr[4:8])
		if ln == 0 || ln > 1<<20 {
			break
		}
		payload := make([]byte, ln)
		if _, err := io.ReadFull(f, payload); err != nil {
			break
		}
		if crc32.ChecksumIEEE(payload) != want {
			break
		}
		rec, err := decodePayload(payload)
		if err != nil {
			break
		}
		out = append(out, rec)
	}
	return out, nil
}

func ReplayAndVerify(path string) ([]Record, error) {
	recs, err := Replay(path)
	if err != nil {
		return nil, err
	}
	for i, rec := range recs {
		if err := Verify(rec); err != nil {
			return nil, fmt.Errorf("recipe: record %d: %w", i, err)
		}
	}
	return recs, nil
}
