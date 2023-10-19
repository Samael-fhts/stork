* [build] slawek

    Disabled CGO to avoid associations with GLIBC and its variants. It allows
    Stork to run on the operating system with the older version of this library
    than installed in the build environment.
    (Gitlab #1201)
