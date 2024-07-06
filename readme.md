## What's this
This program filters the nostk log json file by the date specified in the argument and outputs it to standard output.  

## Install
```
go install github\.com\/mitsugu\/catnostk@&lt;tagname&gt;
```

## Usage
* catnostk -d "2024" -f &lt;nostk log filename&gt;
* catnostk -d "2024/07" -f &lt;nostk log filename&gt;
* catnostk -d "2024/07/06" -f &lt;nostk log filename&gt;

* catnostk -d "2024" &lt; &lt;nostk log filename&gt;
* catnostk -d "2024/07" &lt; &lt;nostk log filename&gt;
* catnostk -d "2024/07/06" &lt; &lt;nostk log filename&gt;
