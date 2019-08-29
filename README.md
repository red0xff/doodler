# doodler
a scrapper for Doodles featured on google.com.

Output is in json format.

# Compilation

`go build doodler.go`

# Help

`doodler -h`

Usage of ./doodler:

	-end string
			Last date to scrap (default "2019/08")
	
	-full
			Query the full format (more informations)

	-hd-image
			Scrap the doodle image in HD
	
	-image
			Scrap the doodle image (not HD resolution)
	
	-output_path string
			Directory where to save the scrapped data (default ".")
	
	-start string
			First date to scrap (default "1998/08")


# Usage Examples

### Scrap basic informations about doodles from August 1998 to this date to the current folder

`doodler`

### Scrap basic informations between February 2016 and Mars 2018

`doodler -start 2016/02 -end 2018/03`

### Scrap all informations (including countries where the doodle was showcased, and other infos)

`doodler -full`

### Scrap both regular and HD versions of doodles (along with basic infos) for the month of May 2016

`doodler -start 2016/05 -end 2016/05 -image -hd-image`
