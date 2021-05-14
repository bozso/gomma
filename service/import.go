package service

import (
	"bufio"
	"fmt"

	"github.com/bozso/gotoolbox/path"

	//"github.com/bozso/emath/geometry"

	"github.com/bozso/gomma/common"
	"github.com/bozso/gomma/date"
	s1 "github.com/bozso/gomma/sentinel1"
)

type SentinelImport struct {
	Output
	Input
	MasterDate date.ShortTime `json:"master_date"`
	Pol        common.Pol     `json:"polarization"`
}

var s1Import = common.Must("S1_import_SLC_from_zipfiles")

func (s *S1Implement) DataImport(si *SentinelImport) (err error) {
	const (
		burst_table = "burst_number_table"
		ziplist     = "ziplist"
	)

	defer si.In.Close()

	zips, err := loadS1(si.In)
	if err != nil {
		return
	}

	if !si.MasterDate.IsSet() {
		return fmt.Errorf("master date is not set")
	}

	masterDate := date.Short.Format(si.MasterDate.Time)
	var master *s1.Zip
	for _, s1zip := range zips {
		if date.Short.Format(s1zip.Date()) == masterDate {
			master = s1zip
		}
	}

	if master == nil {
		return fmt.Errorf("could not find master file, Sentinel 1 zipfile with date '%s' not found", masterDate)
	}

	masterIW, err := master.Info(s.CachePath)
	if err != nil {
		return
	}

	fburst, err := path.New(burst_table).Create()
	if err != nil {
		return
	}
	defer fburst.Close()

	_, err = fmt.Fprintf(fburst, "zipfile: %s\n", master.Path)
	if err != nil {
		return
	}

	nIWs := 0

	for ii, iw := range si.IWs {
		min, max := iw.Min, iw.Max

		if min == 0 && max == 0 {
			continue
		}

		nburst := max - min

		if nburst < 0 {
			return fmt.Errorf(
				"number of burst for IW%d is negative, did you mix up first and last burst numbers?", ii+1)
		}

		IW := masterIW[ii]
		first := IW.bursts[min-1]
		last := IW.bursts[max-1]

		const tpl = "iw%d_number_of_bursts: %d\niw%d_first_burst: %f\niw%d_last_burst: %f\n"
		_, err = fmt.Fprintf(fburst, tpl, ii+1, nburst, ii+1, first, ii+1, last)

		if err != nil {
			return
		}

		nIWs++
	}

	// defer os.Remove(ziplist)

	slcDir, err := s.OutputDir.Join("SLC").Mkdir()
	if err != nil {
		return
	}

	pol, writer := si.Pol, bufio.NewWriter(&si.Out)
	defer si.Out.Close()

	for _, s1zip := range zips {
		// iw, err := s1zip.Info(extInfo)
		date := date.Short.Format(s1zip.Date())
		other := Search(s1zip, zips)

		err = toZiplist(ziplist, s1zip, other)

		if err != nil {
			return fmt.Errorf(
				"could not write zipfiles to zipfile list file '%s'\n%w",
				ziplist, err)
		}

		tab, err := path.New(fmt.Sprintf("%s.%s.SLC_TAB", date, pol)).
			ToValidFile()
		if err != nil {
			return
		}

		slc := s1.FromTabfile()

		for ii := 0; ii < nIWs; ii++ {
			dat := fmt.Sprintf("%s.slc.iw%d", base, ii+1)

			slc.IWs[ii], err = s1.NewIW(dat).Load()
			if err != nil {
				return
			}
		}

		if slc, err = slc.Move(slcDir); err != nil {
			return
		}

		if _, err = writer.WriteString(slc.Tab); err != nil {
			return
		}
	}

	// TODO: save master idx?
	//err = SaveJson(path, meta)
	//
	//if err != nil {
	//    return Handle(err, "failed to write metadata to '%s'", path)
	//}

	return nil
}
