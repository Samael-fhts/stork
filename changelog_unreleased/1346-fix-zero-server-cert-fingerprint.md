[bug] slawek

    Fixed creating a server certificate fingerprint file that contains the zero
    fingerprint if the Stork agent with a version higher than or equal to
    1.15.1 registers in the Stork server 1.15 or less. Prevented the Stork
    agent from running if its certificates are not valid.
    (Gitlab #1346, #1352)
