# pacman-go

playin' with go

## Files
- `maze01.txt`
  - ASCII representation of the pacman maze
  - `#` represents a wall
  - `.` represents a dot
  - `P` represents the player
  - `G` represents the ghosts (enemies)
  - `X` represents the power up pills
- `config.json`
  - emoji settings
- `config_noemoji.json`
  - plain text settings


## Example Commands
`$ go build`
`$ ./pacman-go --help`
`$ ./pacman-go --config-file config_noemoji.json`


## Credits
Many thanks to [@danicat](https://www.youtube.com/watch?v=GH0DlCKTppE)


## todo
- [ ] decouple `main.go`
- [ ] add hyper.sh gif when finished
