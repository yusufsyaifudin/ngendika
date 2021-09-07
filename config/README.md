# config

This folder contains all logic related to binding configuration.

We SHALL NOT have any binding struct from different directories to make sure 
that any changes of config value and key are in one place.

If some library can read the config right from their code (such as Jaeger which read from the env var) 
we encourage to write it again here for the clarity of code.