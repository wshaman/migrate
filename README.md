## YetAnotherMigrate tool

This is just another go migrating tool.
I found go-migrate unusable for current project - 
it's a pain to keep all migrations up-to-date in our team. So this tool allows to run 
any migrations you missing after merging master to your dev branch (sic!)

Also I'd like to run migrations from tests and separate sql folder path is dependant on 
CI tool, and you don't want to mess all this with code. So all migrations are in go files
(you can argue, but I want it.)

* Use Time as a migration ID
* Keep all he migration applied in a table, not the last one
* Keep migrations in go files to be available everywhere