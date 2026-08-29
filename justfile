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
    # --jobs 1: hurl --test runs files in parallel by default, which
    # reorders/interleaves them across workers. The mock IdP's active
    # scenario (see cmd/mockidp/main.go) is shared mutable state across
    # files, so the suite must run one file at a time.
    hurl --test --jobs 1 --file-root tests --retry 5 --retry-interval 1000 tests/hurl/*.hurl; \
    status=$?; \
    docker compose -f docker-compose.test.yml down -v; \
    exit $status
