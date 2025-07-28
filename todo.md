# Steps to MVP

- work with local ollama API
- Dry run funcionality (print results)
- copy - non-destructive run option (default)
-

- add file extension back to new filename - done
- make sure subdirectories map to the same in the output - make sure it isn't just flat - done
- maybe add an option to keep filesystem structure or flatten - done

- add pass ai_model var to openAI client - done
- have different defaults for ai_model depending on whether --local is set - done
-- add model verification on server - done

- make default values for input and output dirs - done

********************************************

- find alternative to jq if I don't have time to implement tessarect or other external cmds (will make installation easier)
- or find way to make Docker container in and out directories intuitive to use. Then it won't matter.
