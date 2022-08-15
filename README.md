# go-gimp-palette
Loader for the GIMP palette format

The format isn't documented, but we can figure it out from [the source code](https://gitlab.gnome.org/GNOME/gimp/-/blob/gimp-2-10/app/core/gimppalette-load.c#L39):

* File contents are UTF-8
* Must have the magic header ```GIMP Palette```
* Following lines are ```Name: <palette name>``` and ```Columns: <no of columns>```
* Comment lines begin with ```#```
* All remaining non-empty lines will be in the format ```65 12 255 color-name``` (the numbers being each rgb byte respectively and the color entry name being optional)