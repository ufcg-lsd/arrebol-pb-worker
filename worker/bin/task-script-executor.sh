#!/bin/bash

# Read the task script file and execute one command at a time, saving your exitcodes in the .ts.ec file.
# Each command executed is written to the .cmds file.
# Use -tsf= or --task_filepath= to input the task file path (Required).
# Use the flag -d or --debug to store .out and .err from execution (Optional).

# This flag does the execution not stop on non-zero exit code commands
set +e

<<<<<<< HEAD
EXECUTION_DIR=/arrebol
=======
WORK_DIR=/arrebol
>>>>>>> master

for i in "$@"
do
	case $i in
	    -tsf=*|--task_filepath=*)
	    __TASK_SCRIPT_FILEPATH="${i#*=}"
	    shift
	    ;;
	    -d|--debug)
	    DEBUG=YES
	    shift
	    ;;
	    *)
	        # unknown option
	    ;;
	esac
done

if [ ! -f "$__TASK_SCRIPT_FILEPATH" ];
then
	echo "$__TASK_SCRIPT_FILEPATH is not a file"
	exit 17
fi

TS_FILENAME=$(basename $__TASK_SCRIPT_FILEPATH)

<<<<<<< HEAD
__EXIT_CODES=$EXECUTION_DIR/$TS_FILENAME.ec
rm $__EXIT_CODES
touch $__EXIT_CODES

__COMMANDS=$EXECUTION_DIR/$TS_FILENAME.cmds
=======
__EXIT_CODES=$WORK_DIR/$TS_FILENAME.ec
rm $__EXIT_CODES
touch $__EXIT_CODES

__COMMANDS=$WORK_DIR/$TS_FILENAME.cmds
>>>>>>> master
rm $__COMMANDS
touch $__COMMANDS

if [ -n "$DEBUG" ];
then
<<<<<<< HEAD
	rm $EXECUTION_DIR/$TS_FILENAME.out
	exec 1> $EXECUTION_DIR/$TS_FILENAME.out
	rm $EXECUTION_DIR/$TS_FILENAME.err
	exec 2> $EXECUTION_DIR/$TS_FILENAME.err
=======
	rm $WORK_DIR/$TS_FILENAME.out
	exec 1> $WORK_DIR/$TS_FILENAME.out
	rm $WORK_DIR/$TS_FILENAME.err
	exec 2> $WORK_DIR/$TS_FILENAME.err
>>>>>>> master
fi

while IFS= read -r __line || [ -n "$__line" ]; do
	set +e
    eval $__line
	  echo $__line >> $__COMMANDS
    echo "$?" >> $__EXIT_CODES
done < $__TASK_SCRIPT_FILEPATH