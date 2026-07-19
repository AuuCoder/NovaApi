-- Grok is the canonical Sub2API platform identifier. Older registration
-- clients used the vendor name "xai", which prevents those accounts from
-- matching Grok groups during scheduling.
UPDATE accounts
SET platform = 'grok', updated_at = NOW()
WHERE LOWER(BTRIM(platform)) = 'xai';

UPDATE groups
SET platform = 'grok', updated_at = NOW()
WHERE LOWER(BTRIM(platform)) = 'xai';
