// Shared example data for the /prototype/hikyo/* marketing prototypes.
// Every claim here is lifted from README.md, PRODUCT.md, DESIGN.md, or the
// shipped docs under src/content/docs. No customers, testimonials, adoption
// figures, pricing, or compliance claims: the repository does not make them.

export const repositoryUrl = 'https://github.com/Hikyo-Org/Hikyo';
export const docsUrl = '/docs/';
export const gettingStartedUrl = '/docs/getting-started/';
export const selfHostingUrl = '/docs/self-hosting/';
export const cliReferenceUrl = '/docs/cli-reference/';
export const coreConceptsUrl = '/docs/core-concepts/';

export type Environment = 'development' | 'staging' | 'production';
export const environments: readonly Environment[] = ['development', 'staging', 'production'];

export type Classification = 'config' | 'secret';

/**
 * Value state vocabulary from DESIGN.md. Every state carries a glyph or word
 * beside its colour so it never depends on colour alone.
 *
 * - set:       explicit config value, shown in plain text
 * - secret:    explicit secret value, always masked
 * - absent:    explicit absence, "· absent"
 * - missing:   presence rule says required, environment is absent: violation
 * - forbidden: presence rule says forbidden, environment is (correctly) absent
 * - draft:     a staged, unpublished change (Δ)
 */
export type CellState =
  | { kind: 'set'; value: string }
  | { kind: 'secret' }
  | { kind: 'absent' }
  | { kind: 'missing' }
  | { kind: 'forbidden' }
  | { kind: 'draft'; value: string };

export interface KeyRow {
  name: string;
  classification: Classification;
  /** Declaration rule, as `hikyo key create --declaration` would receive it. */
  rule: string;
  /** Human wording of the presence rule. */
  presence: string;
  cells: Record<Environment, CellState>;
}

export interface KeyGroup {
  name: string;
  /** All-or-none key group (docs/hierarchy). */
  allOrNone?: boolean;
  keys: KeyRow[];
}

export const project: {
  organisation: string;
  name: string;
  revision: string;
  protectedEnvironments: Environment[];
} = {
  organisation: 'platform',
  name: 'payments-api',
  revision: 'r14',
  protectedEnvironments: ['production'],
};

/** The masked secret placeholder. Prototypes must never render plaintext. */
export const masked = '••••••••••••';

export const matrix: KeyGroup[] = [
  {
    name: 'database',
    keys: [
      {
        name: 'DATABASE_URL',
        classification: 'secret',
        rule: '{"rule":{"type":"url","schemes":["postgres"]}}',
        presence: 'required everywhere',
        cells: { development: { kind: 'secret' }, staging: { kind: 'secret' }, production: { kind: 'secret' } },
      },
      {
        name: 'DATABASE_POOL_MAX',
        classification: 'config',
        rule: '{"rule":{"type":"integer","min":1,"max":200}}',
        presence: 'required everywhere',
        cells: {
          development: { kind: 'set', value: '10' },
          staging: { kind: 'set', value: '20' },
          production: { kind: 'draft', value: '40' },
        },
      },
      {
        name: 'DATABASE_SSLMODE',
        classification: 'config',
        rule: '{"rule":{"type":"enum","values":["disable","require","verify-full"]}}',
        presence: 'required everywhere',
        cells: {
          development: { kind: 'set', value: 'disable' },
          staging: { kind: 'set', value: 'verify-full' },
          production: { kind: 'set', value: 'verify-full' },
        },
      },
    ],
  },
  {
    name: 'payments',
    keys: [
      {
        name: 'STRIPE_SECRET_KEY',
        classification: 'secret',
        rule: '{"rule":{"type":"string","pattern":"^sk_(test|live)_"}}',
        presence: 'required in staging, production',
        cells: { development: { kind: 'absent' }, staging: { kind: 'secret' }, production: { kind: 'secret' } },
      },
      {
        name: 'STRIPE_WEBHOOK_SECRET',
        classification: 'secret',
        rule: '{"rule":{"type":"string","pattern":"^whsec_"}}',
        presence: 'required in staging, production',
        cells: { development: { kind: 'absent' }, staging: { kind: 'secret' }, production: { kind: 'missing' } },
      },
      {
        name: 'PAYMENTS_MOCK_PROVIDER',
        classification: 'config',
        rule: '{"rule":{"type":"boolean"}}',
        presence: 'forbidden in production',
        cells: {
          development: { kind: 'set', value: 'true' },
          staging: { kind: 'set', value: 'false' },
          production: { kind: 'forbidden' },
        },
      },
    ],
  },
  {
    name: 'observability',
    keys: [
      {
        name: 'LOG_LEVEL',
        classification: 'config',
        rule: '{"rule":{"type":"enum","values":["debug","info","warn","error"]}}',
        presence: 'required everywhere',
        cells: {
          development: { kind: 'set', value: 'debug' },
          staging: { kind: 'set', value: 'info' },
          production: { kind: 'set', value: 'info' },
        },
      },
      {
        name: 'SENTRY_DSN',
        classification: 'secret',
        rule: '{"rule":{"type":"url","schemes":["https"]}}',
        presence: 'optional',
        cells: { development: { kind: 'absent' }, staging: { kind: 'secret' }, production: { kind: 'secret' } },
      },
    ],
  },
  {
    name: 'smtp',
    allOrNone: true,
    keys: [
      {
        name: 'SMTP_HOST',
        classification: 'config',
        rule: '{"rule":{"type":"hostname"}}',
        presence: 'group: all or none',
        cells: {
          development: { kind: 'absent' },
          staging: { kind: 'set', value: 'smtp.example.com' },
          production: { kind: 'set', value: 'smtp.example.com' },
        },
      },
      {
        name: 'SMTP_PASSWORD',
        classification: 'secret',
        rule: '{"rule":{"type":"string","minLength":16}}',
        presence: 'group: all or none',
        cells: { development: { kind: 'absent' }, staging: { kind: 'missing' }, production: { kind: 'secret' } },
      },
    ],
  },
];

export const keyCount = matrix.reduce((n, group) => n + group.keys.length, 0);

/** Display text for a cell, following the DESIGN.md state vocabulary. */
export function cellText(cell: CellState): string {
  switch (cell.kind) {
    case 'set':
      return cell.value;
    case 'secret':
      return masked;
    case 'absent':
      return '· absent';
    case 'missing':
      return '! required · absent';
    case 'forbidden':
      return '· absent (forbidden)';
    case 'draft':
      return `Δ ${cell.value}`;
  }
}

/** Accessible description of a cell state. */
export function cellLabel(cell: CellState): string {
  switch (cell.kind) {
    case 'set':
      return `set: ${cell.value}`;
    case 'secret':
      return 'secret set, masked';
    case 'absent':
      return 'absent';
    case 'missing':
      return 'violation: required but absent';
    case 'forbidden':
      return 'absent, forbidden here';
    case 'draft':
      return `draft: ${cell.value}, not yet published`;
  }
}

/** Claims verified against README.md and the shipped docs. */
export const claims = {
  tagline: 'Secrets and configuration you can reason about.',
  positioning:
    'A fully open-source, self-hosted control plane for explicit values across development, staging, and production.',
  license: 'MPL-2.0',
  maturity: 'Active 0.x development. No stable 1.0 release yet.',
  runtime: 'One Go binary with an embedded web UI. SQLite or PostgreSQL.',
  noEnterpriseTier: 'Every capability required to run Hikyo in production is open source. There is no /ee directory.',
  pillars: [
    {
      title: 'Explicit state',
      body: 'Empty never means inherited, unknown, or default. Each environment records set or absent. No environment inherits from another.',
    },
    {
      title: 'Declarations before values',
      body: 'Define config or secret, a validation rule, and a presence rule first. Invalid writes are refused before state changes. A required value cannot be cleared where its rule applies.',
    },
    {
      title: 'Deliberate disclosure',
      body: 'Ordinary reads return presence and metadata. Reveal and copy require reauthentication, remask automatically, and write their own audit events. Replacing a secret never reveals the old one.',
    },
    {
      title: 'One authorization chokepoint',
      body: 'Humans, machine identities, and local break-glass use the same capability-and-scope decision. Unauthorized resources look identical to resources that do not exist.',
    },
    {
      title: 'Self-hosting is the product',
      body: 'One binary, an embedded UI, your database, your root key. Envelope encryption with operator-held root key material.',
    },
  ],
  surfaces: ['Web UI', 'CLI', 'HTTP API'],
  delivery: ['Docker Compose', 'hikyo run', 'Kubernetes operator', 'Forgejo and GitHub Actions adapters'],
  identity: ['OIDC', 'WebAuthn', 'TOTP', 'SAML', 'SCIM'],
  imports: ['.env', 'Kubernetes', 'SOPS', 'Infisical', 'Vault and OpenBao'],
  lifecycle: ['drafts', 'atomic publish', 'snapshots', 'rollback', 'protected environments'],
  audit: 'Append-only audit trail. Events record actor, object, outcome, and decision context, never secret material.',
  scanning: 'Secret scanning on every CLI and API value ingress, and in the browser editor.',
};

export const cli = {
  declare: `hikyo key create --context payments/production \\
  --name STRIPE_WEBHOOK_SECRET --classification secret \\
  --declaration '{"rule":{"type":"string","pattern":"^whsec_"}}'`,
  set: `hikyo values set STRIPE_WEBHOOK_SECRET \\
  --context payments/production --value-file ./whsec.txt`,
  get: `hikyo values get STRIPE_WEBHOOK_SECRET --context payments/production`,
  getMasked: `STRIPE_WEBHOOK_SECRET  secret  set  ${masked}`,
  reveal: `hikyo values get STRIPE_WEBHOOK_SECRET \\
  --context payments/production --reveal --output-file ./whsec`,
  diff: `hikyo values diff --context payments \\
  --left staging --right production`,
  run: `hikyo run --context payments/production -- ./payments-api`,
  server: `HIKYO_DB=sqlite:/var/lib/hikyo/hikyo.db \\
HIKYO_EXTERNAL_ORIGIN=https://hikyo.example.com \\
./hikyo server --listen 0.0.0.0:8443 \\
  --tls-cert-file /etc/hikyo/tls.crt \\
  --tls-key-file /etc/hikyo/tls.key \\
  --root-key-file /etc/hikyo/root.key`,
  dev: `./bin/hikyo server --dev`,
};

/**
 * Reads of a secret that the example matrix records as set in production
 * (STRIPE_SECRET_KEY). Static pages use these so the reveal demo never reads a
 * key the matrix shows as absent.
 */
export const cliRead = {
  get: `hikyo values get STRIPE_SECRET_KEY --context payments/production`,
  getMasked: `STRIPE_SECRET_KEY  secret  set  ${masked}`,
  reveal: `hikyo values get STRIPE_SECRET_KEY \\
  --context payments/production --reveal --output-file ./stripe-key`,
  outputFile: './stripe-key',
};

export interface PrototypeEntry {
  n: number;
  name: string;
  concept: string;
}

export const prototypes: readonly PrototypeEntry[] = [
  { n: 1, name: 'Schematic', concept: 'Annotated technical drawing on linen paper' },
  { n: 2, name: 'Split stage', concept: 'Scroll-driven scenario, CLI and UI side by side' },
  { n: 3, name: 'Departure board', concept: 'The matrix as a split-flap board' },
  { n: 4, name: 'Manifesto', concept: 'Oversized type, hairlines, one accent' },
  { n: 5, name: 'Audit trail', concept: 'The page reads as an append-only log' },
  { n: 6, name: 'Receipt', concept: 'The matrix printed on a thermal receipt, line by line' },
  { n: 7, name: 'Transit map', concept: 'Keys as lines, environments as stations, every stop explicit' },
  { n: 8, name: 'Passport', concept: 'One page per environment, every state stamped in ink' },
  { n: 9, name: 'Inventory', concept: 'The matrix as a game inventory grid, every slot accounted for' },
  { n: 10, name: 'Patch bay', concept: 'Nothing connects unless someone patches it' },
  { n: 11, name: 'Live page, enhanced', concept: 'The current homepage with the matrix resolving on load, callouts, and playing artifacts' },
];
