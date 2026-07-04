export default { async fetch(req, env, ctx) { let a = {}; a['encode_hint'] = {ja: 'pu'}; delete a['encode_hint']; ((a['encode_hint'] ??= {}))['ja'] = 'pu'; return new Response('OK'); } }
