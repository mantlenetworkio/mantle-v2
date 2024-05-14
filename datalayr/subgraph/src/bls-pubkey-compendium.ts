import { BigInt, Address,  Bytes, ethereum, crypto, ByteArray } from "@graphprotocol/graph-ts";
import { log } from '@graphprotocol/graph-ts'
import { NewPubkeyRegistration } from "../generated/BLSPublicKeyCompendium/BLSPublicKeyCompendium";
import { OperatorPublicKeys } from '../generated/schema'

const MAX_UINT32 = BigInt.fromU32(2**32-1)

export function handleNewPubkeyRegistration(event: NewPubkeyRegistration): void {
    log.debug('here', [])
    let operatorId = event.params.operator.toHex()
    let pubkeys = new OperatorPublicKeys(operatorId)

    let pubkeyG1:BigInt[];
    pubkeyG1 = []
    pubkeyG1.push(event.params.pubkeyG1.X)
    pubkeyG1.push(event.params.pubkeyG1.Y)

    pubkeys.pubkeyG1 = pubkeyG1

    let pubkeyG2:BigInt[];
    pubkeyG2 = []
    pubkeyG2.push(event.params.pubkeyG2.X[1])
    pubkeyG2.push(event.params.pubkeyG2.X[0])
    pubkeyG2.push(event.params.pubkeyG2.Y[1])
    pubkeyG2.push(event.params.pubkeyG2.Y[0])

    pubkeys.pubkeyG2 = pubkeyG2

    let pkHex = "0x"
    pkHex += bigIntToHexPad(event.params.pubkeyG1.X, 64)
    pkHex += bigIntToHexPad(event.params.pubkeyG1.Y, 64)

    pubkeys.pubkeyHash = Bytes.fromByteArray(crypto.keccak256(ByteArray.fromHexString(pkHex)))

    pubkeys.save()
}

// f(255, 5) = "000FF"
function bigIntToHexPad(x: BigInt, targetLength: i32): string {
    return x.toHex().substr(2).padStart(targetLength, "0")
}
