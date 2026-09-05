// Run the example/live app first. Requires Playwright on Node's module path.
// GSUI_BROWSER optionally selects an already installed Chromium executable.
const { chromium } = require('playwright');
const assert = require('node:assert/strict');

(async () => {
  const browser = await chromium.launch({headless:true, executablePath:process.env.GSUI_BROWSER || undefined});
  try {
    const context = await browser.newContext();
    await context.addInitScript(() => {
      window.__testSockets=[];
      const Native=window.WebSocket;
      window.WebSocket=class extends Native { constructor(...args){super(...args);window.__testSockets.push(this);} };
    });
    const page = await context.newPage();
    const errors = [];
    page.on('pageerror', error => errors.push(error.message));
    page.on('console', message => { if(message.type()==='error' && message.text().includes('ws exec error')) errors.push(message.text()); });
    const base = process.env.GSUI_URL || 'http://127.0.0.1:1425';
    await page.goto(base);
    await page.waitForFunction(() => window.__ws && __ws.connected());
    await page.evaluate(() => window.navigationSentinel = 'preserved');
    await page.locator('#name').fill('A');
    await page.getByRole('button',{name:'Save',exact:true}).click();
    await page.waitForFunction(() => document.querySelector('[data-gsui-error]'));
    assert.equal(await page.locator('#name').getAttribute('aria-invalid'),'true');
    await page.locator('#name').fill('Alice');
    await page.locator('#name').press('Enter');
    await page.waitForFunction(() => document.getElementById('greeting').textContent === 'Hello, Alice');
    assert.equal(await page.locator('[data-gsui-error]').count(),0);
    await page.getByRole('link',{name:'Counter',exact:true}).click();
    await page.waitForURL('**/counter');
    assert.equal(await page.evaluate(() => window.navigationSentinel),'preserved');
    await page.waitForFunction(() => document.getElementById('count').textContent === '0');
    await page.locator('#draft').fill('Keep this draft');
    await page.locator('#increment').click();
    await page.waitForFunction(() => document.getElementById('count').textContent === '1');
    assert.equal(await page.locator('#draft').inputValue(),'Keep this draft');
    await page.locator('#increment').click();
    await page.waitForFunction(() => document.getElementById('count').textContent === '2');
    await page.getByRole('link',{name:'Change URL without resetting the count'}).click();
    await page.waitForURL('**/counter?tab=details');
    assert.equal(await page.locator('#count').textContent(),'2');
    assert.equal(await page.locator('#draft').inputValue(),'Keep this draft');

    const second = await context.newPage();
    await second.goto(base+'/counter');
    await second.waitForFunction(() => window.__ws && __ws.connected());
    await second.locator('#increment').click();
    await second.waitForFunction(() => document.getElementById('count').textContent === '1');
    assert.equal(await page.locator('#count').textContent(),'2');
    await second.close();

    await page.getByRole('link',{name:'Profile',exact:true}).click();
    await page.waitForURL(base+'/');
    assert.equal(await page.locator('#greeting').textContent(),'Hello, Alice');
    const clock = await page.locator('#clock').textContent();
    await page.waitForFunction(old => document.getElementById('clock').textContent !== old,clock);
    await page.goBack();
    await page.waitForSelector('#count');
    assert.equal(await page.evaluate(() => window.navigationSentinel),'preserved');
    await page.goForward();
    await page.waitForSelector('#greeting');

    await page.getByRole('link',{name:'Products',exact:true}).click();
    await page.waitForURL('**/products');
    assert.equal(await page.locator('#products tbody tr').count(),3);
    await page.getByRole('button',{name:'Load more...',exact:true}).click();
    await page.waitForFunction(()=>document.querySelectorAll('#products tbody tr').length===6);
    assert.match(page.url(),/products.page=2/);
    await page.locator('#products-search').fill('ap');
    await page.locator('#products-search').press('Enter');
    await page.waitForFunction(()=>new URL(location.href).searchParams.get('products.q')==='ap');
    assert.equal(await page.locator('#products tbody tr').count(),3);
    assert.equal(await page.locator('#products-search').inputValue(),'ap');
    await page.getByRole('columnheader',{name:'Name'}).click();
    await page.waitForFunction(()=>new URL(location.href).searchParams.get('products.sort')==='0');
    await page.goBack();
    await page.waitForFunction(()=>document.getElementById('products-search').value==='ap');
    assert.equal(await page.evaluate(()=>window.navigationSentinel),'preserved');
    await page.getByRole('link',{name:'Profile',exact:true}).click();
    await page.waitForSelector('#greeting');

    // A real DOM morph must retain focus/selection and refresh event listeners.
    await page.evaluate(() => {
      const region = document.createElement('div'); region.id='morph-test';
      region.innerHTML='<input id="focused" value="initial"><button id="changing">old</button><span id="tail">tail</span>';
      document.body.append(region);
      const input=region.querySelector('input'); input.value='draft'; input.focus(); input.setSelectionRange(1,3);
      const next=document.createElement('div'); next.id='morph-test';
      next.innerHTML='<span id="tail">moved</span><input id="focused" value="server"><button id="changing">new</button>';
      next.querySelector('button').__gsuiHandlers={click:()=>window.morphedEvent=true};
      __gsuiMorph(region,next);
      if(document.activeElement!==input || input.value!=='draft' || input.selectionStart!==1) throw Error('focus/draft lost');
      region.querySelector('button').click();
      if(!window.morphedEvent) throw Error('event handler not updated');
      const replacement=document.createElement('div');replacement.id='morph-test';
      replacement.innerHTML='<p id="tail">different tag</p><input id="focused" value="reset" data-gsui-reset>';
      __gsuiMorph(region,replacement);
      if(input.value!=='reset')throw Error('explicit input reset failed');
      region.remove();

      const controls=document.createElement('div');
      controls.innerHTML='<select id="choice"><option value="a" selected>A</option><option value="b">B</option></select><input id="check" type="checkbox"><button id="busy" disabled aria-busy="true">Busy</button>';
      document.body.append(controls);
      const fresh=document.createElement('div');
      fresh.innerHTML='<select id="choice"><option value="a">A</option><option value="b" selected>B</option></select><input id="check" type="checkbox" checked><button id="busy">Busy</button>';
      __gsuiMorph(controls,fresh);
      if(controls.querySelector('select').value!=='b'||!controls.querySelector('input').checked)throw Error('pristine controls did not update');
      if(!controls.querySelector('button').disabled)throw Error('morph released another request busy state');
      controls.remove();
    });

    // Reconnect re-arms subscriptions without an HTTP navigation.
    await context.setOffline(true);
    // Chromium's HTTP offline emulation does not reliably close open sockets.
    // Close the actual browser socket to exercise the reconnect path.
    await page.evaluate(()=>window.__testSockets.at(-1).close());
    await page.waitForFunction(() => !__ws.connected());
    await context.setOffline(false);
    await page.waitForFunction(() => __ws.connected());
    const reconnectedClock=await page.locator('#clock').textContent();
    await page.waitForFunction(old=>document.getElementById('clock').textContent!==old,reconnectedClock);
    assert.equal(await page.evaluate(() => window.navigationSentinel),'preserved');
    assert.deepEqual(errors,[]);
    console.log('PASS: forms, navigation/history, view isolation, tables/URL state, keyed morphs, input preservation, subscriptions/reconnect');
    await context.close();
  } finally { await browser.close(); }
})().catch(error => { console.error(error); process.exitCode=1; });
