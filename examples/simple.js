import eth from 'k6/x/substrate';

const client = new eth.Client({});

export default function () {
    const hash = client.getBlockHashLatest();
    console.log(`block => ${hash.hex()}`);

    const block = client.getBlock(hash);
    console.log(JSON.stringify(block.block.extrinsics[0]));
}
