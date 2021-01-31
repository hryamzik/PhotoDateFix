# PhotoDateFix

If time set on your camera was incorrect and you want to adjust times for taken pictures this tool can help. It's designed to adjust time and location in EXIF data. It's recommended to take a picture of watches to use it as a marker.

## Examples

```
photo-date-fix -f test.JPG -t 2018-05-20T23:30:00+07:00 -in data -out out -l 53.0326000,158.6307500 -tz +12:00
photo-date-fix -f input/_marker.JPG -t 2018-05-16T23:16:00+12:00 -in input/ -out output -l 53.0326000,158.6307500 -tz +12:00
photo-date-fix -in input/ -out out_pech -l 53.0326000,158.6307500 -tz +12:00 -d -9h
```
