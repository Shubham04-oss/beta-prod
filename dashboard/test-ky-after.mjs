import ky from 'ky';
const api = ky.create({
  hooks: {
    afterResponse: [
      (args) => {
        console.log("afterResponse args:", Object.keys(args));
      }
    ]
  }
});
api.get('https://example.com').catch(() => {});
