# Hue Entertainment Utility

HEU (Hue Entertainment Utility) is a command-line utility for interacting with the Philips Hue Streaming API.
You provide channel and colour information, and it handles the details.

## Configuration

Configuration can be read either using environmnent variables or via a configuration file located within the user's configuration directory (for linux this will usually be `~/.config/heu/heu.toml`) or within the same directory as the heu executable.
For a list of configuration keys, see [.env.example](.env.example) and [heu.toml.example](heu.toml.example).
To populate the configuration, follow the setup process within the [hue entertainment documentation](https://developers.meethue.com/develop/hue-entertainment/hue-entertainment-api/).

## Usage

HEU does not have the ability to create new entertainment configurations, so you'll need to do that within the hue app.
To query existing entertainment configurations, use `heu list`

```bash
$ heu list
{ "errors" [], "data": [...] }
```

To initiate an entertainment streaming session, use `heu stream`

```bash
# Stream from file `commands.txt`, reading 3 command lines per second, looped
$ heu stream -f commands.txt -k 3 -l

# Stream from stdin, with no delay reading commands
$ heu stream
```

### Streaming command format

The stream command format is meant to be simple to construct by hand.
Each line contains a set of `;`-separated commands, each of which can be used to set a colour for a single channel, or for every channel.
If it were defined by a grammar it would be something like the following

```EBNF
command line = { command chunk, ";" }, [ command chunk ] ;
command chunk = [ whitespace ], [ channel id ], " ", hex colour, [ whitespace ] ;
channel id = "<int>" ;
hex colour = "#<6-digit hex colour>" ;
whitespace = "<any unicode whitespace>" ;
```

so a line containing `0 #FF0000; 1 #00FF00` would send commands to turn channel 0 red and channel 1 green.
A line containing `#0000FF` would send commands to turn all channels blue.
See [commands.txt.example](commands.txt.example) for an example command file.
