// Headless checks of the form controls in uictl.js: check boxes, text areas,
// tabs, sliders, combo boxes and edit boxes.
//   node web/uictl2.test.mjs

import assert from 'node:assert/strict';
import { UIData, buildWindow } from './uiwin.js';
import { wrapText, sliderValue, sliderX } from './uictl.js';

const n = (t, body) => {
  if (Array.isArray(body)) return { t, c: body };
  if (body && typeof body === 'object') return { t, c: Object.entries(body).map(([k, v]) => n(k, v)) };
  return { t, x: String(body) };
};

function fakeDoc() {
  return {
    createElement(tag) {
      const e = {
        tag, style: {}, dataset: {}, children: [], listeners: {},
        appendChild(c) { this.children.push(c); return c; },
        append(...cs) { cs.forEach((c) => this.children.push(c)); },
        addEventListener(t, f) { this.listeners[t] = f; },
      };
      let text = '';
      Object.defineProperty(e, 'textContent', { get: () => text, set: (v) => { text = v; if (v === '') e.children = []; } });
      return e;
    },
  };
}

// wrapText: words break at spaces, newlines are kept, long words are cut.
{
  const measure = (t) => t.length * 5;
  assert.deepEqual(wrapText(measure, 'one two three', 100), ['one two three']);
  assert.deepEqual(wrapText(measure, 'one two three', 40), ['one', 'two', 'three'].slice(0, 0).concat(['one two', 'three']));
  assert.deepEqual(wrapText(measure, 'a\nb', 100), ['a', 'b'], 'a newline starts a line');
  assert.deepEqual(wrapText(measure, 'a\n\nb', 100), ['a', '', 'b'], 'a blank line stays');
  assert.deepEqual(wrapText(measure, 'abcdefghij', 25), ['abcde', 'fghij'], 'a word too long for a line is cut');
  assert.deepEqual(wrapText(measure, '', 50), [''], 'no text is one empty line');
  assert.deepEqual(wrapText(measure, 'hello   world', 100), ['hello   world'], 'spaces inside a line are kept');
  console.log('ok: text wraps at spaces and keeps its lines');
}

// A slider's value and place.
{
  assert.equal(sliderValue(0, 150, 51), 0);
  assert.equal(sliderValue(150, 150, 51), 50);
  assert.equal(sliderValue(75, 150, 51), 25);
  assert.equal(sliderValue(-20, 150, 51), 0, 'before the track is the first value');
  assert.equal(sliderValue(999, 150, 51), 50, 'past the track is the last');
  assert.equal(sliderX(25, 150, 51), 75);
  assert.equal(sliderX(80, 150, 51), 150, 'a value past the end sits at the end');
  assert.equal(sliderX(5, 100, 1), 0, 'one tick has no travel');
  console.log('ok: a slider maps pixels to values and back');
}

const layout = {
  skin: 'test',
  sizes: { 'atlantis/a.tga': [256, 256] },
  bitmap: {},
  elements: [
    n('Texture', { Name: 'page1', File: 'atlantis/a.tga' }),
    n('CheckBoxTemplate', [n('Name', 'chk'), n('Size', { X: 12, Y: 12 }), n('Font', [n('Name', 'brit9'), n('TextOffset', { X: 17, Y: 0 })]),
      n('Texture', [n('TextureName', 'page1'), n('Normal', { X: 76, Y: 241 }), n('Pressed', { X: 63, Y: 241 }), n('NormalHighlit', { X: 102, Y: 241 }), n('PressedHighlit', { X: 89, Y: 241 }), n('Disabled', { X: 76, Y: 241 })])]),
    n('TextAreaTemplate', [n('Name', 'ta'), n('Font', [n('Name', 'arial11')]), n('TextOffset', { X: 10, Y: 5 })]),
    n('ButtonTemplate', [n('Name', 'tabbtn'), n('Size', { X: 50, Y: 16 })]),
    n('TabsTemplate', [n('Name', 'tabs'), n('TabButtonTemplate', 'tabbtn'), n('TabXOffset', 5)]),
    n('ImageAreaTemplate', [n('Name', 'mark'), n('TextureName', 'page1'), n('Size', { X: 13, Y: 7 }), n('TopLeft', { X: 165, Y: 247 })]),
    n('HorizontalSliderTemplate', [n('Name', 'slide'), n('ImageAreaTemplate', 'mark'), n('Height', 13)]),
    n('ComboBoxTemplate', [n('Name', 'combo')]),
    n('EditBoxTemplate', [n('Name', 'edit'), n('BackgroundTemplate', 'none')]),
    n('WindowTemplate', [
      n('Name', 'forms'), n('WindowId', 'Forms'), n('Width', 300), n('Height', 200),
      n('CheckBoxDef', [n('TemplateName', 'chk'), n('ControlId', '1000'), n('AdapterName', 'pet_aggro'), n('Data', 'Aggro'), n('LabelWidth', 46), n('Position', { X: 8, Y: 25 })]),
      n('TextAreaDef', [n('TemplateName', 'ta'), n('ControlId', '1001'), n('AdapterName', 'story'), n('Width', 200), n('Height', 60), n('Position', { X: 10, Y: 50 })]),
      n('TabsDef', [n('TemplateName', 'tabs'), n('ControlId', '1002'), n('Width', 150), n('Height', 80), n('Position', { X: 0, Y: 120 }),
        n('Tab', [n('Id', '1'), n('Name', 'One')]), n('Tab', [n('Id', '2'), n('Name', 'Two')]),
        n('TabControl', [n('TabId', '1'), n('ControlId', '1003')]), n('TabControl', [n('TabId', '2'), n('ControlId', '1004')])]),
      n('HorizontalSliderDef', [n('TemplateName', 'slide'), n('ControlId', '1003'), n('Adaptername', 'spec_level0'), n('Width', 113), n('Numticks', 51), n('Position', { X: 170, Y: 20 })]),
      n('ComboBoxDef', [n('Templatename', 'combo'), n('ControlId', '1004'), n('Position', { X: 170, Y: 60 })]),
      n('EditBoxDef', [n('TemplateName', 'edit'), n('ControlId', '1005'), n('Width', 100), n('Height', 16), n('Position', { X: 170, Y: 100 })]),
    ]),
  ],
};
const data = new UIData(layout, 'data/ui/test');
const win = buildWindow(fakeDoc(), data, 'Forms');

// A check box toggles, reports and can be set, relabelled and disabled.
{
  const cb = win.check('pet_aggro');
  assert.ok(cb && win.check('1000') === cb, 'found by adapter and by control id');
  assert.equal(win.ctl['1000'].style.left, '8px');
  assert.equal(cb.get(), false);
  const seen = [];
  cb.onChange((on) => seen.push(on));
  const art = win.ctl['1000'].children[0];
  const off = art.style.backgroundPosition;
  win.ctl['1000'].listeners.click();
  assert.equal(cb.get(), true);
  assert.deepEqual(seen, [true]);
  assert.notEqual(art.style.backgroundPosition, off, 'ticked shows the pressed art');
  win.ctl['1000'].listeners.click();
  assert.deepEqual(seen, [true, false]);
  cb.set(true);
  assert.equal(cb.get(), true);
  assert.equal(seen.length, 2, 'set does not report: only a click does');
  cb.disable();
  win.ctl['1000'].listeners.click();
  assert.equal(cb.get(), true, 'a disabled box ignores clicks');
  assert.equal(win.ctl['1000'].children[1].children.length, 1, 'the caption is drawn');
  cb.setLabel('');
  assert.equal(win.ctl['1000'].children[1].children.length, 0);
  console.log('ok: a check box toggles, reports and can be disabled');
}

// A text area wraps what it is given and scrolls.
{
  const area = win.area('story');
  assert.ok(area && win.area('1001') === area);
  area.setText(Array.from({ length: 30 }, (_, i) => 'Line number ' + i).join('\n'));
  assert.equal(area.lines.length, 30);
  assert.ok(area.view >= 2 && area.view < 30);
  area.scrollTo(5);
  assert.equal(area.top, 5);
  area.scrollTo(999);
  assert.equal(area.top, 30 - area.view, 'stops at the end');
  area.setText('short');
  assert.deepEqual(area.lines, ['short']);
  assert.equal(area.top, 0, 'new text starts at the top');
  console.log('ok: a text area wraps and scrolls');
}

// Tabs show the controls of the chosen tab.
{
  const tabs = win.tabs('1002');
  assert.deepEqual(tabs.tabs.map((t) => t.name), ['One', 'Two']);
  assert.equal(tabs.selected, '1', 'the first tab is chosen once the window is drawn');
  assert.equal(win.ctl['1003'].style.display, '');
  assert.equal(win.ctl['1004'].style.display, 'none', 'the other tab is hidden');
  const changes = [];
  tabs.onChange((id) => changes.push(id));
  tabs.select('2');
  assert.equal(win.ctl['1003'].style.display, 'none');
  assert.equal(win.ctl['1004'].style.display, '');
  assert.deepEqual(changes, ['2']);
  tabs.select('1');
  console.log('ok: tabs show their own controls');
}

// A slider takes and reports values.
{
  const sl = win.slider('spec_level0'); // the layout spells it Adaptername
  assert.ok(sl && win.slider('1003') === sl);
  assert.equal(sl.ticks, 51);
  sl.set(25);
  assert.equal(sl.get(), 25);
  sl.set(900);
  assert.equal(sl.get(), 50, 'clamped');
  const heard = [];
  sl.onChange((v) => heard.push(v));
  const track = win.ctl['1003'];
  track.getBoundingClientRect = () => ({ left: 100 });
  track.listeners.pointerdown({ clientX: 100 + 6 + 50 }); // the middle of the 100-px track, past the half indicator
  assert.equal(sl.get(), 25);
  track.listeners.pointermove({ clientX: 100 + 6 + 100 });
  assert.equal(sl.get(), 50);
  track.listeners.pointerup();
  track.listeners.pointermove({ clientX: 100 });
  assert.equal(sl.get(), 50, 'a pointer that is not down does not drag');
  assert.deepEqual(heard, [25, 50]);
  console.log('ok: a slider takes values from the pointer');
}

// A combo box steps through its options.
{
  const cb = win.combo('1004');
  assert.equal(cb.selected, -1);
  cb.setOptions(['Low', 'Medium', 'High']);
  assert.equal(cb.value, 'Low');
  const heard = [];
  cb.onChange((i, o) => heard.push([i, o]));
  cb.select(2);
  assert.equal(cb.value, 'High');
  cb.el.children[0].listeners.click();
  assert.equal(cb.value, 'Low', 'a click steps on, round to the start');
  assert.deepEqual(heard, [[2, 'High'], [0, 'Low']]);
  cb.setOptions([]);
  assert.equal(cb.selected, -1);
  console.log('ok: a combo box steps through its options');
}

// An edit box holds text and says when Enter is pressed.
{
  const ed = win.edit('1005');
  ed.set('hello');
  assert.equal(ed.value, 'hello');
  const lines = [];
  ed.onEnter((t) => lines.push(t));
  ed.input.listeners.keydown({ key: 'a' });
  ed.input.listeners.keydown({ key: 'Enter' });
  assert.deepEqual(lines, ['hello']);
  console.log('ok: an edit box reports its line on Enter');
}

// A rebuild keeps the form controls working.
{
  win.resize(320, 210);
  assert.ok(win.check('pet_aggro') && win.tabs('1002').selected === '1');
  console.log('ok: resizing keeps the form controls');
}
