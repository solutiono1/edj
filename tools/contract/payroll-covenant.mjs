// Builds a PrivatePayrollBatch instance from a payroll CSV. Capacity is four
// recipients per covenant because CashScript 0.14 has no array parameters.
// Larger payrolls must be split into batches.

import { readFileSync } from 'node:fs'
import { Contract, MockNetworkProvider } from 'cashscript'
import { Address, NETWORK } from 'cashaddr'
import { hex } from '@scure/base'

const artifact = JSON.parse(readFileSync(new URL('./PrivatePayrollBatch.json', import.meta.url), 'utf8'))
const ZERO_PKH = '00'.repeat(20)
const DEFAULT_PAYER_PKH = '1d0f172a0e0f4b6edf9b4e5e4bd4f4fd67f6ad3c'

export function parsePkh(address) {
  try {
    const decoded = Address(NETWORK).decode(address)
    if (decoded.type !== 'pkh') throw new Error('not a P2PKH address')
    return hex.encode(decoded.hash)
  } catch (error) {
    throw new Error('Invalid cashaddr: ' + address)
  }
}

export function buildBatch(rows, payerPkh = DEFAULT_PAYER_PKH, refundAfter = 900000n) {
  if (!Array.isArray(rows) || rows.length < 1 || rows.length > 4) {
    throw new Error('A batch needs 1 to 4 recipients')
  }
  const payees = []
  const amounts = []
  for (const row of rows) {
    if (!row.address || !Number.isInteger(row.amount) || row.amount <= 0) {
      throw new Error('Each row needs an address and a positive integer amount')
    }
    payees.push(parsePkh(row.address))
    amounts.push(BigInt(row.amount))
  }
  while (payees.length < 4) {
    payees.push(ZERO_PKH)
    amounts.push(0n)
  }
  const total = amounts.reduce((sum, value) => sum + value, 0n)
  const args = [
    ...payees, ...amounts, BigInt(rows.length), total, payerPkh, refundAfter
  ]
  const contract = new Contract(artifact, args, { provider: new MockNetworkProvider() })
  return {
    address: contract.address,
    count: rows.length,
    total,
    args: args.map((value) => value.toString()),
    note: 'Fund this address with total plus a fee. The covenant permits only the fixed outputs or a refund after the height.'
  }
}

const isMain = process.argv[1] && import.meta.url.endsWith(process.argv[1].split('/').pop())
if (isMain) {
  const file = process.argv[2]
  if (!file) {
    console.error('Usage: node payroll-covenant.mjs <payroll.csv>')
    process.exit(1)
  }
  const lines = readFileSync(file, 'utf8').trim().split(/\r?\n/)
  const header = lines.shift()?.split(',').map((x) => x.trim().toLowerCase())
  if (header?.join(',') !== 'name,address,amount') {
    console.error('Expected header name,address,amount')
    process.exit(1)
  }
  const rows = lines.map((line) => {
    const [name, address, amount] = line.split(',').map((x) => x.trim())
    return { name, address, amount: Number(amount) }
  })
  console.log(JSON.stringify(buildBatch(rows), null, 2))
}
