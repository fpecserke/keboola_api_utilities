# Keboola data download tool 

Took me about 4 min to download it all (it is parallel).

## Download

Run `go run main.go -m -t` to download everything.
You can run any combination of flags:
- `-m`: Downloads metadata (buckets, workspaces, components, tables)
- `-t`: Downloads table details (column types, nullable, size, etc...)
Run `go run main.go -h` for more detailed help.

## Search

There are few shell scripts that use `jq` to extract data from `.json` files (it needs to be installed).
They print links to any transformation (old legacy/new legacy/version 2) that have the 
searched text in the name of any inputs.

Some transformations from new legacy and old legacy transformations are overlapping.

### Making the script executable
If you don't have permissions to run this program, make it executable:
`chmod +x ./transformations.sh`

### Running the Shell Script
Run them like this:
`./transformations.sh $SEARCH_PHRASE`

Feel free to change the code or the script, just don't hold me accountable ;)
