import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { Contract, MockNetworkProvider } from 'cashscript'
import { Address, NETWORK } from 'cashaddr'
import { hex } from '@scure/base'

const artifact = JSON.parse(readFileSync(new URL('./PrivatePayroll.json', import.meta.url), 'utf8'))
const payeePkh = 'f5bf48b397dae70be82b3cca4793f8eb2b6cdac9'
const payerPkh = '1d0f172a0e0f4b6edf9b4e5e4bd4f4fd67f6ad3c'
const amount = 25_000n
const refundAfter = 900_000n

const provider = new MockNetworkProvider()
const contract = new Contract(artifact, [payeePkh, payerPkh, amount, refundAfter], { provider })

const pkhAddress = (pkh) => Address(NETWORK).encode({
  format: 'cashaddr',
  type: 'pkh',
  hash: hex.decode(pkh)
})

const payeeAddress = pkhAddress(payeePkh)
const payerAddress = pkhAddress(payerPkh)
const utxo = {
  txid: 'a1'.repeat(32),
  vout: 0,
  satoshis: amount + 1_000n
}

provider.addUtxo(contract.address, utxo)

const pay = () => contract.functions.pay()
  .from(utxo)
  .to(payeeAddress, amount)
  .withoutChange()

const payDebug = await pay().debug()
assert.ok(payDebug, 'pay covenant evaluates successfully')
const payTx = await pay().build()
assert.equal(typeof payTx, 'string')
assert.ok(payTx.includes(payeePkh))
console.log('PASS pay covenant evaluates and constrains the output')

const refund = () => contract.functions.refund()
  .from(utxo)
  .withTime(Number(refundAfter))
  .to(payerAddress, amount)
  .withoutChange()

const refundDebug = await refund().debug()
assert.ok(refundDebug, 'refund covenant evaluates successfully')
const refundTx = await refund().build()
assert.ok(refundTx.includes(payerPkh))
console.log('PASS refund covenant evaluates and constrains the output')

let rejected = false
try {
  await contract.functions.pay()
    .from(utxo)
    .to(payerAddress, amount)
    .withoutChange()
    .debug()
} catch (error) {
  rejected = true
}
assert.equal(rejected, true, 'payer address is rejected by the pay covenant')
console.log('PASS pay covenant rejects a different recipient')

console.log('artifact', artifact.contractName, contract.address)
