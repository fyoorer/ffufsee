# ffufsee

Browse Ffuf's JSON output as a neatly formatted HTML report in the comfort of your browser.

But... but... Ffuf already has HTML output flag available!? I know, I am using the same code 😁

ffufsee was created for specific use cases:
1. You have the JSON file generated by Ffuf and you want to generate HTML report. As far as I know, Ffuf does not have this option. 
2. You are like me and run distributed Ffuf using [ShadowClone](https://github.com/fyoorer/ShadowClone) so generating hundreds of HTML reports does not make sense. JSON files can be combined easily!

## Installation & Usage
```bash
git clone https://github.com/fyoorer/ffufsee.git
cd ffufsee
go run main.go ~/path/to/your/ffuf-output.json
```

Or you can build a binary and execute from anywhere
```bash
go build .
cp ffufsee /usr/local/bin/
ffufsee ~/path/to/your/ffuf-output.json
``` 

Open browser and visit `http://localhost:5505`

## Note!
The Ffuf JSON file needs to be created by Ffuf v2.0.0 (current version). JSON files generated by older Ffuf versions are not supported, sorry!

