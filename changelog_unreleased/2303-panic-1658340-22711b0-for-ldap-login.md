[bug] slawek

    The LDAP hook identifies now users by a unique identifier rather
    than a distinguished name (DN). Old identifiers stored in the
    database are updated on login.
    (Gitlab #2303)
