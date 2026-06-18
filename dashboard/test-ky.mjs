import ky from 'ky';
const api = ky.create({
  hooks: {
    beforeRequest: [
      (request, options) => {
        console.log("request is:", typeof request, request ? Object.keys(request) : 'null');
        console.log("has headers?", request && request.headers ? true : false);
      }
    ]
  }
});
api.get('https://example.com').catch(e => console.error(e.message));
