import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { Contract, MockNetworkProvider } from 'cashscript'
import { Address, NETWORK } from 'cashaddr'
import { hex } from '@scure/base'

const artifact = JSON.parse(readFileSync(new URL('./PrivatePayrollBatch.json', import.meta.url), 'utf8'))
const payerPkh = '1d0f172a0e0f4b6edf9b4e5e4bd4f4fd67f6ad3c'
const payees = [
  'f5bf48b397dae70be82b3cca4793f8eb2b6cdac9',
  'a1b2c3d4e5f60718293a4b5c6d7e8f9001a2b3c4',
  '0123456789abcdef0123456789abcdef01234567',
  'fedcba9876543210fedcba9876543210fedcba98'
]
const amounts = [25_000n, 15_000n, 9_000n, 4_000n]
const refundAfter = 900_000n

const pkhAddress = (pkh) => Address(NETWORK).encode({
  format: 'cashaddr',
  type: 'pkh',
  hash: hex.decode(pkh)
})
const addresses = payees.map(pkhAddress)
const payerAddress = pkhAddress(payerPkh)

function setup(count) {
  const total = amounts.slice(0, count).reduce((a, b) => a + b, 0n)
  const provider = new MockNetworkProvider()
  const contract = new Contract(artifact, [
    ...payees, ...amounts, BigInt(count), total, payerPkh, refundAfter
  ], { provider })
  const utxo = { txid: 'b1'.repeat(32), vout: 0, satoshis: total + 1_000n }
  provider.addUtxo(contract.address, utxo)
  return { contract, utxo, total }
}

const payTx = (ctx, count, outputs) => {
  let tx = ctx.contract.functions.pay().from(ctx.utxo)
  const list = outputs || Array.from({ length: count }, (_, i) => [addresses[i], amounts[i]])
  for (const [address, amount] of list) tx = tx.to(address, amount)
  return tx.withoutChange()
}

const refundTx = (ctx, time, address, amount) => ctx.contract.functions.refund()
  .from(ctx.utxo)
  .withTime(Number(time))
  .to(address, amount)
  .withoutChange()

const ctx2 = setup(2)
assert.ok(await payTx(ctx2, 2).debug(), 'two-recipient pay covenant evaluates')
console.log('PASS batch pay with two recipients')

const ctx4 = setup(4)
assert.ok(await payTx(ctx4, 4).debug(), 'four-recipient pay covenant evaluates')
console.log('PASS batch pay with four recipients')

const rejects = async (label, build) => {
  let rejected = false
  try {
    await build().debug()
  } catch (error) {
    rejected = true
  }
  assert.equal(rejected, true, label + ' was accepted')
  console.log('PASS ' + label)
}

await rejects('wrong recipient is rejected', () => payTx(ctx2, 2, [
  [payerAddress, amounts[0]], [addresses[1], amounts[1]]
]))

await rejects('wrong amount is rejected', () => payTx(ctx2, 2, [
  [addresses[0], amounts[0] + 1n], [addresses[1], amounts[1]]
]))

await rejects('wrong output count is rejected', () => payTx(ctx2, 1))

assert.ok(await refundTx(ctx2, refundAfter, payerAddress, ctx2.total).debug(), 'refund evaluates')
console.log('PASS batch refund after refund height')

await rejects('refund before the height is rejected', () =>
  refundTx(ctx2, refundAfter - 1n, payerAddress, ctx2.total))

await rejects('refund to another address is rejected', () =>
  refundTx(ctx2, refundAfter, addresses[0], ctx2.total))

console.log('artifact', artifact.contractName, ctx2.contract.address)
