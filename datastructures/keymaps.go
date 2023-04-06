package datastructures

 const (
	KeyMap string = `Everywhere
--
q           Closes the application
< >         Arrow Keys, Navigate between the main windows
∧ v         Arroy Keys, Scroll through the lists of the main windows
x           Open KeyMap View
ctrl + x    Close KeyMapView

Repositories - View
--
Enter       Load Jobs for selected Repository
f           Filter the repositories list

Job - View
--
Enter       Load Runs for selected Job
f           Filter the job list - TBD
L           Open Launch Window with the default config for this job

Runs - View
--
L           Open Launch Window with the config from the selected run


Filter - View
--
Enter       Apply Filter to the list of items of the view from where the filter has been launched from, or if empty, brings you back to the view

LaunchConfigEditor - View
--
Enter       (Verifies TBD) and Launches a Run of the Job with the displayed config
???         Closes the Launch Window, Changes are not saved
???         Toggle comment in selected line
Arrow Keys  Navigation (TBD)

`
 )