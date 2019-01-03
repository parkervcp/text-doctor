# text-doctor
A stream overlay text manager using gist files or google sheets.

This is a simple golang application to read a google sheet and write its contents to a single text file for use with OBS as a text file interface.

This will allow for a group to edit and update the contents of a stream overlay.

The config allows for specifying which columns and rows for data to be read from.

```
{
	"sheet": {
        "sheet_id": "", ## The ID for the google sheet Ex. "https://docs.google.com/spreadsheets/d/<this_part_here>/edit#gid=0"
        "sheet_table":"", ## The table name to use in the sheet Ex. "Class Data"
        "sheet_cells_start": "", ## The first column and cell to start reading from Ex. "A2" (first column, second cell)
        "sheet_cell_end": "", ## The last column to read data from Ex. "E" (data in f and after will not be read)
		"sheet_columns": [ ## The columns to write out into the text file (starts at 0 [0,1,2,3,4])
			"0",
			"4"
		],
		"sheet_refresh_interval": 60 ## How often the gsheet should be polled for new data (in seconds) min 60.
	},
	"file": {
		"file_location": "overlay.txt" ## what file to write to.
	}
}
```