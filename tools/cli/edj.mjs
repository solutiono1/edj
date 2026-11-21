#!/usr/bin/env node
import { readFileSync } from 'node:fs'
import { EdjClient } from '../../sdk/js/edj.js'

const baseUrl = process.env.EDJ_URL || 'http://localhost:8080'
const token = process.env.EDJ_TOKEN
if (!token) {
  console.error('Set EDJ_TOKEN to your session token')
  process.exit(1)
}

const edj = new EdjClient({ baseUrl, token })
const [command, ...args] = process.argv.slice(2)

const show = (value) => console.log(JSON.stringify(value, null, 2))

try {
  if (command === 'me') {
    show(await edj.me())
  } else if (command === 'portfolio') {
    show(await edj.portfolio())
  } else if (command === 'links' && args[0] === 'create') {
    const amount = Number(args[1])
    const tokenId = args[2] && args[2] !== '-' ? args[2] : null
    const note = args.slice(3).join(' ')
    show(await edj.createLink({ amount, token: tokenId, note }))
  } else if (command === 'payroll' && args[0] === 'upload') {
    const csv = readFileSync(args[1], 'utf8')
    show(await edj.submitPayroll(csv, args[1].split('/').pop()))
  } else if (command === 'proofs' && args[0] === 'create') {
    show(await edj.createProof({
      subject: args[1],
      threshold: Number(args[2]),
      expiresInDays: args[3] ? Number(args[3]) : 30
    }))
  } else {
    console.log('Commands:')
    console.log('  me')
    console.log('  portfolio')
    console.log('  links create <amount> [token|-] [note]')
    console.log('  payroll upload <csv-file>')
    console.log('  proofs create <subject> <threshold> [days]')
  }
} catch (error) {
  console.error(error.message)
  process.exit(1)
}
