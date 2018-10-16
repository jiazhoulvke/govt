# govt #

Generate sudoku-like thumbnails from videos.

## Dependencies ##

- [FFmpeg](https://www.ffmpeg.org/)
- [FFProbe](https://www.ffmpeg.org/ffprobe.html)

### Mac OS X ##

```bash
brew install ffmpeg
```

## Installation ##

### Build From Source Code ###

```bash
go get -v github.com/jiazhoulvke/govt
```

## Usage ##

```
Usage of govt:
  -allow_exts string
        video format (default ".avi,.rmvb,.rm,.asf,.divx,.mpg,.mpeg,.mpe,.wmv,.mp4,.mkv,.vob,.mov")
  -max_height int
        picture max height (default 1000)
  -max_width int
        picture max width (default 1000)
  -save_dir string
        save picture at(if null then use video's path)
  -save_name string
        picture name(if null then use video's name)
  -start_at int
        start at (second) (default 10)
```

## Examples ##

### Folder ###

```bash
govt -allow_exts ".avi,.mov,.mkv" ~/Movies/blender
```

### Multiple Files ###

```bash
govt ~/Movies/blender/big_buck_bunny_720p_h264.mov ~/Movies/blender/Sintel.2010.720p.mkv
```

### Parameters ###

```bash
govt -start_at 60 -max_height 1280 -max_width 1280 ~/Movies/blender/big_buck_bunny_720p_h264.mov
```

## ScreenShot ##

![screenshot1](https://github.com/jiazoulvke/govt/raw/master/screenshots/big_buck_bunny_720p_h264.jpg)

![screenshot2](https://github.com/jiazoulvke/govt/raw/master/screenshots/Sintel.2010.720p.jpg)

## License ##

govt is licensed under GPLv3 license.