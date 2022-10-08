import eth from 'k6/x/substrate';

const client = new eth.Client({});

export default function () {
    const block = client.getBlockHashLatest();
    console.log(`block => ${block.hex()}`);
}
