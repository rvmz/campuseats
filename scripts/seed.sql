-- Session 3. Run by hand after goose up. This is not a migration:
--   psql "$DATABASE_URL" -f scripts/seed.sql
-- Fixed ids so later curls match. See docs/session-03.md.

insert into users (id, email, role) values
    ('11111111-1111-1111-1111-111111111111', 'ada@campus.edu', 'student'),
    ('22222222-2222-2222-2222-222222222222', 'kitchen@campus.edu', 'kitchen');

insert into restaurants (id, name, campus) values
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'North Hall Noodles', 'north');

insert into menu_items (id, restaurant_id, name, price_cents) values
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Miso ramen', 950),
    ('cccccccc-cccc-cccc-cccc-cccccccccccc', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Rice bowl', 800);
