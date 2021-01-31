package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/tajtiattila/metadata/exif"
)

var (
	imageRegex        = regexp.MustCompile(`(?i)(.*)\.(jpe?g)$`)
	srcPathString     = flag.String("in", ".", "source directory, default is current one")
	dstPathString     = flag.String("out", ".", "destination directory, default is current one")
	nameSuffix        = flag.String("s", "", "if source and destination directories match output files should have a suffix defined")
	location          = flag.String("l", "0.0,0.0", "location where pictures were taken, should match desired timezone")
	infoMode          = flag.Bool("i", false, "info mode, print time delta and exit")
	fileNameString    = flag.String("f", "", "file to check date against")
	timeString        = flag.String("t", "", "time in RFC3339 format, for example 2018-05-21T23:30:00+12:00")
	offsetString      = flag.String("tz", "", "timezone as GMT offset, example: +03:00 stands for MSK")
	delta             = flag.Duration("d", 0, "time delta to be applied, ignored if -f and -t flags are set")
	defaultNameSuffix = "_date_fixed"
	dryRun            = flag.Bool("c", false, "dry run")
)

func CheckError(err error) {
	if err != nil {
		log.Println(err)
	}
}

func CheckErrorFatal(err error) {
	if err != nil {
		panic(fmt.Sprintf("Error: %s\n", err))
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Get time delta: %s -f filename -t timestring\n", os.Args[0])
	flag.Usage()
}

func parceLocationString(locationString string) (lat, lon float64) {
	splitted := strings.Split(locationString, ",")
	if len(splitted) != 2 {
		CheckErrorFatal(
			errors.New(
				fmt.Sprintf("Can't parce location '%s'", locationString),
			),
		)
	}
	lat, err := strconv.ParseFloat(splitted[0], 64)
	if err != nil {
		CheckErrorFatal(
			errors.New(
				fmt.Sprintf("Can't parce location '%s': %#v", locationString, err),
			),
		)
	}

	lon, err = strconv.ParseFloat(splitted[1], 64)
	if err != nil {
		CheckErrorFatal(
			errors.New(
				fmt.Sprintf("Can't parce location '%s': %#v", locationString, err),
			),
		)
	}
	return
}

func getDateTimeDelta(metadata *exif.Exif, realTimeString string) time.Duration {
	realTime, err := time.Parse(time.RFC3339, realTimeString)
	CheckErrorFatal(err)
	dateTag, ok := metadata.DateTime()
	if !ok {
		CheckErrorFatal(errors.New("Can't get date from file exif"))
	}
	log.Printf("EXIF date %s is %s than %s\n", dateTag.String(), humanize.RelTime(dateTag, realTime, "earlier", "later"), realTime.String())
	return realTime.Sub(dateTag)
}

func getExifForFile(fileName string) *exif.Exif {
	f, err := os.Open(fileName)
	CheckErrorFatal(err)
	defer f.Close()
	exifReader := bufio.NewReader(f)
	metadata, err := exif.Decode(exifReader)
	CheckErrorFatal(err)
	return metadata
}

func processFile(src, dst string, delta time.Duration, dstZone string, setGPS bool, lat, lon float64, location *time.Location) {

	metadata := getExifForFile(src)

	dateTag, ok := metadata.DateTime()
	if !ok {
		CheckErrorFatal(errors.New("Can't get date from file exif"))
	}

	newTime := dateTag.In(location).Add(delta)

	log.Printf("%s => %s: %s => %s%s\n", src, dst, dateTag.String(), newTime.String(),
		func() string {
			if setGPS {
				return fmt.Sprintf(", location: %f,%f", lat, lon)
			}
			return ""
		}(),
	)

	if *dryRun {
		return
	}

	metadata.SetDateTime(newTime)

	if setGPS {
		gps := exif.GPSInfo{
			Time: newTime,
			Lat:  lat,
			Long: lon,
		}
		metadata.SetGPSInfo(gps)
	}

	f, err := os.Open(src)
	CheckErrorFatal(err)
	defer f.Close()
	reader := bufio.NewReader(f)
	of, err := os.Create(dst)
	CheckErrorFatal(err)
	writer := bufio.NewWriter(of)
	defer of.Close()
	err = exif.Copy(writer, reader, metadata)
	CheckErrorFatal(err)
	writer.Flush()
}

func main() {
	log.Printf("%s", "Starting...")
	flag.Parse()
	if *fileNameString != "" {
		metadata := getExifForFile(*fileNameString)
		*delta = getDateTimeDelta(metadata, *timeString)
		log.Printf("Delta is %dns\n", *delta)
	}
	if *infoMode {
		fmt.Printf("%dns\n", *delta)
		os.Exit(0)
	}
	if *srcPathString == *dstPathString && *nameSuffix == "" {
		*nameSuffix = defaultNameSuffix
		log.Printf("Source and destination directories match but no suffix is defined, using default one: %s", *nameSuffix)
	}

	if _, err := os.Stat(*dstPathString); os.IsNotExist(err) {
		log.Printf("Creating output directory '%s'", *dstPathString)
		err := os.Mkdir(*dstPathString, os.ModePerm)
		CheckErrorFatal(err)
	}
	lat, lon := parceLocationString(*location)
	setGPS := *location != ""
	//														 2018-05-21T23:30:00.000+12:00
	offsetTime, err := time.Parse(time.RFC3339, fmt.Sprintf("2006-01-02T15:04:05%s", *offsetString))
	CheckErrorFatal(err)
	log.Printf("Going to set timezone to GMT%s", offsetTime.Format("-07:00"))

	srcPathInfo, err := os.Stat(*srcPathString)
	CheckErrorFatal(err)

	err = filepath.Walk(*srcPathString, func(path string, f os.FileInfo, err error) error {
		// log.Printf("Checking %s, %s", *srcPathString, filepath.Join(*srcPathString, f.Name()))
		if err != nil {
			log.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if !f.IsDir() {
			if imageRegex.MatchString(f.Name()) {
				src := filepath.Join(*srcPathString, f.Name())
				dst := filepath.Join(*dstPathString, imageRegex.ReplaceAllString(f.Name(), fmt.Sprintf(`${1}%s.${2}`, *nameSuffix)))
				processFile(src, dst, *delta, "", setGPS, lat, lon, offsetTime.Location())
			} else {
				log.Printf("Skipping %s", f.Name())
			}
		}

		if f.IsDir() && !os.SameFile(srcPathInfo, f) {
			log.Printf("skipping a dir without errors: %+v \n", f.Name())
			return filepath.SkipDir
		}
		return nil
	})
	CheckErrorFatal(err)
}

// func setExifGPSDateTime(x *exif.Exif, t time.Time) {
// 	if t.IsZero() {
// 		x.Set(exiftag.GPSDateStamp, nil)
// 		x.Set(exiftag.GPSTimeStamp, nil)
// 		return
// 	}

// 	// GPS time is always UTC.
// 	t = t.UTC()

// 	x.Set(exiftag.GPSDateStamp, exif.Ascii(t.Format("2006:01:02")))

// 	h, m, s := t.Clock()

// 	sn, sd := uint32(s), uint32(1)

// 	// use microsecond precision to avoid uint32 overflow
// 	if us := t.Nanosecond() / 1000; us != 0 {
// 		sn, sd = sn*1e6+uint32(us), 1e6
// 	}

// 	x.Set(exiftag.GPSTimeStamp, exif.Rational{uint32(h), 1, uint32(m), 1, sn, sd})
// }
