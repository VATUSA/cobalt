all:

test-integration:
    docker build -t localhost/cobalt .
    docker build -t localhost/cobalt-mockidp -f Dockerfile.mockidp .
    # `seed` is a one-shot container; docker compose's --wait treats any
    # exited container (even exit 0) as a failed wait, so it's run
    # separately via `run` (which blocks and surfaces its own exit code)
    # after the long-lived services are up.
    docker compose -f docker-compose.test.yml up -d --wait mysql mock-idp cobalt
    docker compose -f docker-compose.test.yml run --rm seed
    # solo_cert.hurl needs "soon" dates relative to whenever this actually
    # runs (expires validation is checked against the real clock), computed
    # here rather than hardcoded so the suite doesn't rot as calendar time
    # passes it by.
    #
    # --jobs 1: hurl --test runs files in parallel by default, which
    # reorders/interleaves them across workers. The mock IdP's active
    # scenario (see cmd/mockidp/main.go) is shared mutable state across
    # files, so the suite must run one file at a time.
    #
    # Every line below has to stay one continuous backslash-continued shell
    # invocation (a bare, non-continued comment line would start a fresh
    # invocation under just's default line-per-shell-call behavior, losing
    # the EXPIRES_* variables before hurl ever sees them) -- that's why the
    # explanation is up here instead of interleaved below.
    EXPIRES_SOON=$(date -u -d '+10 days' +%Y-%m-%d); \
    EXPIRES_LATER=$(date -u -d '+12 days' +%Y-%m-%d); \
    hurl --test --jobs 1 --file-root tests --retry 5 --retry-interval 1000 \
        --variable expires_soon=$EXPIRES_SOON --variable expires_later=$EXPIRES_LATER \
        tests/hurl/*.hurl; \
    status=$?; \
    docker compose -f docker-compose.test.yml down -v; \
    exit $status
