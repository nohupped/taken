# taken
Uses inotify to scan user homedirectories for new deb files, and uses aptly to add it to debian repo.


Todo: 

>Add checks for supporting rsync.

Why: 

>With rsync, inotify reports a notify.Create on a tmp file that rsync creates, and reports notify.InCloseWrite after completion >of writing to that tmp file, and then it just moves the tmp file, so it only reports notify.Create, the second time. This >program actions on notify.InCloseWrite to confirm that the file is completely written to disk before doing any action.  
>Possible workaround for now: use the --inplace flag with rsync to write in place to the original file.
