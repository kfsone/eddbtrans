Elite Dangerous Market Data Translation
=======================================

This is a proof-of-concept implementation of a Go-Menacing gomschema translator which takes
input from EDDB files and converts them into the Go-Meancing protobuffer format.

All data is ingested in parallel. A facility called "Daycare" is used to cope with things being
loaded out-of-order and filter out cases such as stations for which no system is known.

TODO:

The daycare function needs cleaning up, and the parse station daftly fiddly now that I added
daycare, because I've inlined too much functionality rather than specializing.

