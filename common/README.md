# About this file

This file contains the data shared between the backend and frontend
sides.

Because they are strongly separated, and the toolkits don't allow access
to the files above the project directories (i.e., "backend" and
"webui"), the build system copies the files to the target locations. The
copied files should be read-only. The developers should modify only the
original files in this directory; their changes will be reflected in the
copies.

## Updating the generated code for standard DHCP option definitions

DHCP standard options have well-known formats defined in the RFCs. 
Stork backend and frontend are aware of these formats and use them 
to parse option data received from Kea and send updated data to Kea.
When new options are standardized, the Stork code must be updated to
recognize the new options. In that case, a developer should define new
options in the files located in this directory.

You can force generating the definition files by:

```console
$ rake gen:std_option_defs
```

If you provide any changes in the standard DHCP option definition JSONs,
you should format them by calling the below command to avoid warnings
from the UI linter:

```console
$ rake fmt:ui SCOPE="../common/*.json"
```
