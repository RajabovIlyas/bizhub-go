"use strict";(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[536],{87536:function(a,b,c){c.d(b,{Dq:function(){return am},Qr:function(){return E},cI:function(){return aG}});var d=c(67294),e=a=>"checkbox"===a.type,f=a=>a instanceof Date,g=a=>null==a;let h=a=>"object"==typeof a;var i=a=>!g(a)&&!Array.isArray(a)&&h(a)&&!f(a),j=a=>i(a)&&a.target?e(a.target)?a.target.checked:a.target.value:a,k=a=>a.substring(0,a.search(/\.\d+(\.|$)/))||a,l=(a,b)=>a.has(k(b)),m=a=>Array.isArray(a)?a.filter(Boolean):[],n=a=>void 0===a,o=(a,b,c)=>{if(!b||!i(a))return c;let d=m(b.split(/[,[\].]+?/)).reduce((a,b)=>g(a)?a:a[b],a);return n(d)||d===a?n(a[b])?c:a[b]:d};let p={BLUR:"blur",FOCUS_OUT:"focusout",CHANGE:"change"},q={onBlur:"onBlur",onChange:"onChange",onSubmit:"onSubmit",onTouched:"onTouched",all:"all"},r={max:"max",min:"min",maxLength:"maxLength",minLength:"minLength",pattern:"pattern",required:"required",validate:"validate"},s=d.createContext(null),t=()=>d.useContext(s);var u=(a,b,c,d=!0)=>{let e={};for(let f in a)Object.defineProperty(e,f,{get(){let e=f;return b[e]!==q.all&&(b[e]=!d||q.all),c&&(c[e]=!0),a[e]}});return e},v=a=>i(a)&&!Object.keys(a).length,w=(a,b,c)=>{let{name:d,...e}=a;return v(e)||Object.keys(e).length>=Object.keys(b).length||Object.keys(e).find(a=>b[a]===(!c||q.all))},x=a=>Array.isArray(a)?a:[a],y=(a,b,c)=>c&&b?a===b:!a||!b||a===b||x(a).some(a=>a&&(a.startsWith(b)||b.startsWith(a)));function z(a){let b=d.useRef(a);b.current=a,d.useEffect(()=>{let c=a=>{a&&a.unsubscribe()},d=!a.disabled&&b.current.subject.subscribe({next:b.current.callback});return()=>c(d)},[a.disabled])}var A=a=>"string"==typeof a,B=(a,b,c,d)=>{let e=Array.isArray(a);return A(a)?(d&&b.watch.add(a),o(c,a)):e?a.map(a=>(d&&b.watch.add(a),o(c,a))):(d&&(b.watchAll=!0),c)},C=a=>"function"==typeof a,D=a=>{for(let b in a)if(C(a[b]))return!0;return!1};let E=a=>a.render(function(a){let b=t(),{name:c,control:e=b.control,shouldUnregister:f}=a,g=l(e._names.array,c),h=function(a){let b=t(),{control:c=b.control,name:e,defaultValue:f,disabled:g,exact:h}=a||{},j=d.useRef(e);j.current=e;let k=d.useCallback(a=>{if(y(j.current,a.name,h)){let b=B(j.current,c._names,a.values||c._formValues);m(n(j.current)||i(b)&&!D(b)?{...b}:Array.isArray(b)?[...b]:n(b)?f:b)}},[c,h,f]);z({disabled:g,subject:c._subjects.watch,callback:k});let[l,m]=d.useState(n(f)?c._getWatch(e):f);return d.useEffect(()=>{c._removeUnmounted()}),l}({control:e,name:c,defaultValue:o(e._formValues,c,o(e._defaultValues,c,a.defaultValue)),exact:!0}),k=function(a){let b=t(),{control:c=b.control,disabled:e,name:f,exact:g}=a||{},[h,i]=d.useState(c._formState),j=d.useRef({isDirty:!1,dirtyFields:!1,touchedFields:!1,isValidating:!1,isValid:!1,errors:!1}),k=d.useRef(f),l=d.useRef(!0);k.current=f;let m=d.useCallback(a=>l.current&&y(k.current,a.name,g)&&w(a,j.current)&&i({...c._formState,...a}),[c,g]);return z({disabled:e,callback:m,subject:c._subjects.state}),d.useEffect(()=>(l.current=!0,()=>{l.current=!1}),[]),u(h,c._proxyFormState,j.current,!1)}({control:e,name:c}),m=d.useRef(e.register(c,{...a.rules,value:h}));return d.useEffect(()=>{let a=(a,b)=>{let c=o(e._fields,a);c&&(c._f.mount=b)};return a(c,!0),()=>{let b=e._options.shouldUnregister||f;(g?b&&!e._stateFlags.action:b)?e.unregister(c):a(c,!1)}},[c,e,g,f]),{field:{name:c,value:h,onChange:d.useCallback(a=>{m.current.onChange({target:{value:j(a),name:c},type:p.CHANGE})},[c]),onBlur:d.useCallback(()=>{m.current.onBlur({target:{value:o(e._formValues,c),name:c},type:p.BLUR})},[c,e]),ref:d.useCallback(a=>{let b=o(e._fields,c);a&&b&&a.focus&&(b._f.ref={focus:()=>a.focus(),select:()=>a.select(),setCustomValidity:b=>a.setCustomValidity(b),reportValidity:()=>a.reportValidity()})},[c,e._fields])},formState:k,fieldState:Object.defineProperties({},{invalid:{get:()=>!!o(k.errors,c)},isDirty:{get:()=>!!o(k.dirtyFields,c)},isTouched:{get:()=>!!o(k.touchedFields,c)},error:{get:()=>o(k.errors,c)}})}}(a));var F=(a,b,c,d,e)=>b?{...c[a],types:{...c[a]&&c[a].types?c[a].types:{},[d]:e||!0}}:{},G=a=>/^\w*$/.test(a),H=a=>m(a.replace(/["|']|\]/g,"").split(/\.|\[/));function I(a,b,c){let d=-1,e=G(b)?[b]:H(b),f=e.length,g=f-1;for(;++d<f;){let h=e[d],j=c;if(d!==g){let k=a[h];j=i(k)||Array.isArray(k)?k:isNaN(+e[d+1])?{}:[]}a[h]=j,a=a[h]}return a}let J=(a,b,c)=>{for(let d of c||Object.keys(a)){let e=o(a,d);if(e){let{_f:f,...g}=e;if(f&&b(f.name)){if(f.ref.focus&&n(f.ref.focus()))break;if(f.refs){f.refs[0].focus();break}}else i(g)&&J(g,b)}}};var K=()=>{let a="undefined"==typeof performance?Date.now():1e3*performance.now();return"xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g,b=>{let c=(16*Math.random()+a)%16|0;return("x"==b?c:3&c|8).toString(16)})},L=(a,b,c={})=>c.shouldFocus||n(c.shouldFocus)?c.focusName||`${a}.${n(c.focusIndex)?b:c.focusIndex}.`:"",M=(a,b,c)=>!c&&(b.watchAll||b.watch.has(a)||[...b.watch].some(b=>a.startsWith(b)&&/^\.\w+/.test(a.slice(b.length)))),N=(a,b,c)=>{let d=m(o(a,c));return I(d,"root",b[c]),I(a,c,d),a},O=a=>"boolean"==typeof a,P=a=>"file"===a.type,Q=a=>A(a)||d.isValidElement(a),R=a=>"radio"===a.type,S=a=>a instanceof RegExp;let T={value:!1,isValid:!1},U={value:!0,isValid:!0};var V=a=>{if(Array.isArray(a)){if(a.length>1){let b=a.filter(a=>a&&a.checked&&!a.disabled).map(a=>a.value);return{value:b,isValid:!!b.length}}return a[0].checked&&!a[0].disabled?a[0].attributes&&!n(a[0].attributes.value)?n(a[0].value)||""===a[0].value?U:{value:a[0].value,isValid:!0}:U:T}return T};let W={isValid:!1,value:null};var X=a=>Array.isArray(a)?a.reduce((a,b)=>b&&b.checked&&!b.disabled?{isValid:!0,value:b.value}:a,W):W;function Y(a,b,c="validate"){if(Q(a)||Array.isArray(a)&&a.every(Q)||O(a)&&!a)return{type:c,message:Q(a)?a:"",ref:b}}var Z=a=>i(a)&&!S(a)?a:{value:a,message:""},$=async(a,b,c,d,f)=>{let{ref:h,refs:j,required:k,maxLength:l,minLength:m,min:n,max:o,pattern:p,validate:q,name:s,valueAsNumber:t,mount:u,disabled:w}=a._f;if(!u||w)return{};let x=j?j[0]:h,y=a=>{d&&x.reportValidity&&(x.setCustomValidity(O(a)?"":a||" "),x.reportValidity())},z={},B=R(h),D=e(h),E=(t||P(h))&&!h.value||""===b||Array.isArray(b)&&!b.length,G=F.bind(null,s,c,z),H=(a,b,c,d=r.maxLength,e=r.minLength)=>{let f=a?b:c;z[s]={type:a?d:e,message:f,ref:h,...G(a?d:e,f)}};if(f?!Array.isArray(b)||!b.length:k&&(!(B||D)&&(E||g(b))||O(b)&&!b||D&&!V(j).isValid||B&&!X(j).isValid)){let{value:I,message:J}=Q(k)?{value:!!k,message:k}:Z(k);if(I&&(z[s]={type:r.required,message:J,ref:x,...G(r.required,J)},!c))return y(J),z}if(!E&&(!g(n)||!g(o))){let K,L,M=Z(o),N=Z(n);if(g(b)||isNaN(b)){let T=h.valueAsDate||new Date(b);A(M.value)&&(K=T>new Date(M.value)),A(N.value)&&(L=T<new Date(N.value))}else{let U=h.valueAsNumber||+b;g(M.value)||(K=U>M.value),g(N.value)||(L=U<N.value)}if((K||L)&&(H(!!K,M.message,N.message,r.max,r.min),!c))return y(z[s].message),z}if((l||m)&&!E&&(A(b)||f&&Array.isArray(b))){let W=Z(l),$=Z(m),_=!g(W.value)&&b.length>W.value,aa=!g($.value)&&b.length<$.value;if((_||aa)&&(H(_,W.message,$.message),!c))return y(z[s].message),z}if(p&&!E&&A(b)){let{value:ab,message:ac}=Z(p);if(S(ab)&&!b.match(ab)&&(z[s]={type:r.pattern,message:ac,ref:h,...G(r.pattern,ac)},!c))return y(ac),z}if(q){if(C(q)){let ad=await q(b),ae=Y(ad,x);if(ae&&(z[s]={...ae,...G(r.validate,ae.message)},!c))return y(ae.message),z}else if(i(q)){let af={};for(let ag in q){if(!v(af)&&!c)break;let ah=Y(await q[ag](b),x,ag);ah&&(af={...ah,...G(ag,ah.message)},y(ah.message),c&&(z[s]=af))}if(!v(af)&&(z[s]={ref:x,...af},!c))return z}}return y(!0),z};function _(a,b){return[...a,...x(b)]}var aa="undefined"!=typeof window&& void 0!==window.HTMLElement&&"undefined"!=typeof document;function ab(a){let b,c=Array.isArray(a);if(a instanceof Date)b=new Date(a);else if(a instanceof Set)b=new Set(a);else if(!(!(aa&&(a instanceof Blob||a instanceof FileList))&&(c||i(a))))return a;else for(let d in b=c?[]:{},a){if(C(a[d])){b=a;break}b[d]=ab(a[d])}return b}var ac=a=>Array.isArray(a)?a.map(()=>void 0):void 0,ad=a=>({isOnSubmit:!a||a===q.onSubmit,isOnBlur:a===q.onBlur,isOnChange:a===q.onChange,isOnAll:a===q.all,isOnTouch:a===q.onTouched});function ae(a,b,c){return[...a.slice(0,b),...x(c),...a.slice(b),]}var af=(a,b,c)=>Array.isArray(a)?(n(a[c])&&(a[c]=void 0),a.splice(c,0,a.splice(b,1)[0]),a):[];function ag(a,b){return[...x(b),...x(a)]}var ah=(a,b)=>n(b)?[]:function(a,b){let c=0,d=[...a];for(let e of b)d.splice(e-c,1),c++;return m(d).length?d:[]}(a,x(b).sort((a,b)=>a-b)),ai=(a,b,c)=>{var d;a[b]=(d=a[c],a[c]=a[b],d)};function aj(a){for(let b in a)if(!n(a[b]))return!1;return!0}function ak(a,b){let c=G(b)?[b]:H(b),d=1==c.length?a:function(a,b){let c=b.slice(0,-1).length,d=0;for(;d<c;)a=n(a)?d++:a[b[d++]];return a}(a,c),e=c[c.length-1],f;d&&delete d[e];for(let g=0;g<c.slice(0,-1).length;g++){let h=-1,j,k=c.slice(0,-(g+1)),l=k.length-1;for(g>0&&(f=a);++h<k.length;){let m=k[h];j=j?j[m]:a[m],l===h&&(i(j)&&v(j)||Array.isArray(j)&&aj(j))&&(f?delete f[m]:delete a[m]),f=j}}return a}var al=(a,b,c)=>(a[b]=c,a);function am(a){let b=t(),{control:c=b.control,name:e,keyName:f="id",shouldUnregister:g}=a,[h,i]=d.useState(c._getFieldArray(e)),j=d.useRef(c._getFieldArray(e).map(K)),k=d.useRef(h),l=d.useRef(e),m=d.useRef(!1);l.current=e,k.current=h,c._names.array.add(e),a.rules&&c.register(e,a.rules);let n=d.useCallback(({values:a,name:b})=>{if(b===l.current||!b){let c=o(a,l.current,[]);i(c),j.current=c.map(K)}},[]);z({callback:n,subject:c._subjects.array});let p=d.useCallback(a=>{m.current=!0,c._updateFieldArray(e,a)},[c,e]),r=(a,b)=>{let d=x(ab(a)),f=_(c._getFieldArray(e),d);c._names.focus=L(e,f.length-1,b),j.current=_(j.current,d.map(K)),p(f),i(f),c._updateFieldArray(e,f,_,{argA:ac(a)})},s=(a,b)=>{let d=x(ab(a)),f=ag(c._getFieldArray(e),d);c._names.focus=L(e,0,b),j.current=ag(j.current,d.map(K)),p(f),i(f),c._updateFieldArray(e,f,ag,{argA:ac(a)})},u=a=>{let b=ah(c._getFieldArray(e),a);j.current=ah(j.current,a),p(b),i(b),c._updateFieldArray(e,b,ah,{argA:a})},w=(a,b,d)=>{let f=x(ab(b)),g=ae(c._getFieldArray(e),a,f);c._names.focus=L(e,a,d),j.current=ae(j.current,a,f.map(K)),p(g),i(g),c._updateFieldArray(e,g,ae,{argA:a,argB:ac(b)})},y=(a,b)=>{let d=c._getFieldArray(e);ai(d,a,b),ai(j.current,a,b),p(d),i(d),c._updateFieldArray(e,d,ai,{argA:a,argB:b},!1)},A=(a,b)=>{let d=c._getFieldArray(e);af(d,a,b),af(j.current,a,b),p(d),i(d),c._updateFieldArray(e,d,af,{argA:a,argB:b},!1)},B=(a,b)=>{let d=ab(b),f=al(c._getFieldArray(e),a,d);j.current=[...f].map((b,c)=>b&&c!==a?j.current[c]:K()),p(f),i([...f]),c._updateFieldArray(e,f,al,{argA:a,argB:d},!0,!1)},C=a=>{let b=x(ab(a));j.current=b.map(K),p([...b]),i([...b]),c._updateFieldArray(e,[...b],a=>a,{},!0,!1)};return d.useEffect(()=>{if(c._stateFlags.action=!1,M(e,c._names)&&c._subjects.state.next({}),m.current){if(c._options.resolver)c._executeSchema([e]).then(a=>{let b=o(a.errors,e),d=o(c._formState.errors,e);(d?!b&&d.type:b&&b.type)&&(b?I(c._formState.errors,e,b):ak(c._formState.errors,e),c._subjects.state.next({errors:c._formState.errors}))});else{let a=o(c._fields,e),b=ad(c._options.mode);(!b.isOnSubmit||c._formState.isSubmitted)&&a&&a._f&&$(a,o(c._formValues,e),c._options.criteriaMode===q.all,c._options.shouldUseNativeValidation,!0).then(a=>!v(a)&&c._subjects.state.next({errors:N(c._formState.errors,a,e)}))}}c._subjects.watch.next({name:e,values:c._formValues}),c._names.focus&&J(c._fields,a=>a.startsWith(c._names.focus)),c._names.focus="",c._proxyFormState.isValid&&c._updateValid()},[h,e,c]),d.useEffect(()=>(o(c._formValues,e)||c._updateFieldArray(e),()=>{(c._options.shouldUnregister||g)&&c.unregister(e)}),[e,c,f,g]),{swap:d.useCallback(y,[p,e,c]),move:d.useCallback(A,[p,e,c]),prepend:d.useCallback(s,[p,e,c]),append:d.useCallback(r,[p,e,c]),remove:d.useCallback(u,[p,e,c]),insert:d.useCallback(w,[p,e,c]),update:d.useCallback(B,[p,e,c]),replace:d.useCallback(C,[p,e,c]),fields:d.useMemo(()=>h.map((a,b)=>({...a,[f]:j.current[b]||K()})),[h,f])}}function an(){let a=[],b=b=>{for(let c of a)c.next(b)},c=b=>(a.push(b),{unsubscribe(){a=a.filter(a=>a!==b)}}),d=()=>{a=[]};return{get observers(){return a},next:b,subscribe:c,unsubscribe:d}}var ao=a=>g(a)||!h(a);function ap(a,b){if(ao(a)||ao(b))return a===b;if(f(a)&&f(b))return a.getTime()===b.getTime();let c=Object.keys(a),d=Object.keys(b);if(c.length!==d.length)return!1;for(let e of c){let g=a[e];if(!d.includes(e))return!1;if("ref"!==e){let h=b[e];if(f(g)&&f(h)||i(g)&&i(h)||Array.isArray(g)&&Array.isArray(h)?!ap(g,h):g!==h)return!1}}return!0}var aq=a=>{let b=a?a.ownerDocument:0,c=b&&b.defaultView?b.defaultView.HTMLElement:HTMLElement;return a instanceof c},ar=a=>"select-multiple"===a.type,as=a=>R(a)||e(a),at=a=>aq(a)&&a.isConnected;function au(a,b={}){let c=Array.isArray(a);if(i(a)||c)for(let d in a)Array.isArray(a[d])||i(a[d])&&!D(a[d])?(b[d]=Array.isArray(a[d])?[]:{},au(a[d],b[d])):g(a[d])||(b[d]=!0);return b}function av(a,b,c){let d=Array.isArray(a);if(i(a)||d)for(let e in a)Array.isArray(a[e])||i(a[e])&&!D(a[e])?n(b)||ao(c[e])?c[e]=Array.isArray(a[e])?au(a[e],[]):{...au(a[e])}:av(a[e],g(b)?{}:b[e],c[e]):c[e]=!ap(a[e],b[e]);return c}var aw=(a,b)=>av(a,b,au(b)),ax=(a,{valueAsNumber:b,valueAsDate:c,setValueAs:d})=>n(a)?a:b?""===a||g(a)?NaN:+a:c&&A(a)?new Date(a):d?d(a):a;function ay(a){let b=a.ref;return(a.refs?a.refs.every(a=>a.disabled):b.disabled)?void 0:P(b)?b.files:R(b)?X(a.refs).value:ar(b)?[...b.selectedOptions].map(({value:a})=>a):e(b)?V(a.refs).value:ax(n(b.value)?a.ref.value:b.value,a)}var az=(a,b,c,d)=>{let e={};for(let f of a){let g=o(b,f);g&&I(e,f,g._f)}return{criteriaMode:c,names:[...a],fields:e,shouldUseNativeValidation:d}},aA=a=>n(a)?void 0:S(a)?a.source:i(a)?S(a.value)?a.value.source:a.value:a,aB=a=>a.mount&&(a.required||a.min||a.max||a.maxLength||a.minLength||a.pattern||a.validate);function aC(a,b,c){let d=o(a,c);if(d||G(c))return{error:d,name:c};let e=c.split(".");for(;e.length;){let f=e.join("."),g=o(b,f),h=o(a,f);if(g&&!Array.isArray(g)&&c!==f)break;if(h&&h.type)return{name:f,error:h};e.pop()}return{name:c}}var aD=(a,b,c,d,e)=>!e.isOnAll&&(!c&&e.isOnTouch?!(b||a):(c?d.isOnBlur:e.isOnBlur)?!a:(c?!d.isOnChange:!e.isOnChange)||a),aE=(a,b)=>!m(o(a,b)).length&&ak(a,b);let aF={mode:q.onSubmit,reValidateMode:q.onChange,shouldFocusError:!0};function aG(a={}){let b=d.useRef(),[c,h]=d.useState({isDirty:!1,isValidating:!1,dirtyFields:{},isSubmitted:!1,submitCount:0,touchedFields:{},isSubmitting:!1,isSubmitSuccessful:!1,isValid:!1,errors:{}});b.current?b.current.control._options=a:b.current={...function(a={}){let b={...aF,...a},c={isDirty:!1,isValidating:!1,dirtyFields:{},isSubmitted:!1,submitCount:0,touchedFields:{},isSubmitting:!1,isSubmitSuccessful:!1,isValid:!1,errors:{}},d={},h=ab(b.defaultValues)||{},i=b.shouldUnregister?{}:ab(h),k={action:!1,mount:!1,watch:!1},r={mount:new Set,unMount:new Set,array:new Set,watch:new Set},s,t=0,u={},w={isDirty:!1,dirtyFields:!1,touchedFields:!1,isValidating:!1,isValid:!1,errors:!1},y={watch:an(),array:an(),state:an()},z=ad(b.mode),D=ad(b.reValidateMode),E=b.criteriaMode===q.all,F=a=>b=>{clearTimeout(t),t=window.setTimeout(a,b)},G=async a=>{let e=!1;return w.isValid&&(e=b.resolver?v((await S()).errors):await U(d,!0),a||e===c.isValid||(c.isValid=e,y.state.next({isValid:e}))),e},H=(a,b=[],e,f,g=!0,j=!0)=>{if(f&&e){if(k.action=!0,j&&Array.isArray(o(d,a))){let l=e(o(d,a),f.argA,f.argB);g&&I(d,a,l)}if(w.errors&&j&&Array.isArray(o(c.errors,a))){let m=e(o(c.errors,a),f.argA,f.argB);g&&I(c.errors,a,m),aE(c.errors,a)}if(w.touchedFields&&j&&Array.isArray(o(c.touchedFields,a))){let n=e(o(c.touchedFields,a),f.argA,f.argB);g&&I(c.touchedFields,a,n)}w.dirtyFields&&(c.dirtyFields=aw(h,i)),y.state.next({isDirty:W(a,b),dirtyFields:c.dirtyFields,errors:c.errors,isValid:c.isValid})}else I(i,a,b)},K=(a,b)=>{I(c.errors,a,b),y.state.next({errors:c.errors})},L=(a,b,c,e)=>{let f=o(d,a);if(f){let g=o(i,a,n(c)?o(h,a):c);n(g)||e&&e.defaultChecked||b?I(i,a,b?g:ay(f._f)):Z(a,g),k.mount&&G()}},Q=(a,b,d,e,f)=>{let g=!1,i={name:a},j=o(c.touchedFields,a);if(w.isDirty){let k=c.isDirty;c.isDirty=i.isDirty=W(),g=k!==i.isDirty}if(w.dirtyFields&&(!d||e)){let l=o(c.dirtyFields,a),m=ap(o(h,a),b);m?ak(c.dirtyFields,a):I(c.dirtyFields,a,!0),i.dirtyFields=c.dirtyFields,g=g||l!==o(c.dirtyFields,a)}return d&&!j&&(I(c.touchedFields,a,d),i.touchedFields=c.touchedFields,g=g||w.touchedFields&&j!==d),g&&f&&y.state.next(i),g?i:{}},R=async(b,d,e,f)=>{let g=o(c.errors,b),h=w.isValid&&c.isValid!==d;if(a.delayError&&e?(s=F(()=>K(b,e)))(a.delayError):(clearTimeout(t),s=null,e?I(c.errors,b,e):ak(c.errors,b)),(e?!ap(g,e):g)||!v(f)||h){let i={...f,...h?{isValid:d}:{},errors:c.errors,name:b};c={...c,...i},y.state.next(i)}u[b]--,w.isValidating&&!Object.values(u).some(a=>a)&&(y.state.next({isValidating:!1}),u={})},S=async a=>b.resolver?await b.resolver({...i},b.context,az(a||r.mount,d,b.criteriaMode,b.shouldUseNativeValidation)):{},T=async a=>{let{errors:b}=await S();if(a)for(let d of a){let e=o(b,d);e?I(c.errors,d,e):ak(c.errors,d)}else c.errors=b;return b},U=async(a,d,e={valid:!0})=>{for(let f in a){let g=a[f];if(g){let{_f:h,...j}=g;if(h){let k=r.array.has(h.name),l=await $(g,o(i,h.name),E,b.shouldUseNativeValidation,k);if(l[h.name]&&(e.valid=!1,d))break;d||(o(l,h.name)?k?N(c.errors,l,h.name):I(c.errors,h.name,l[h.name]):ak(c.errors,h.name))}j&&await U(j,d,e)}}return e.valid},V=()=>{for(let a of r.unMount){let b=o(d,a);b&&(b._f.refs?b._f.refs.every(a=>!at(a)):!at(b._f.ref))&&am(a)}r.unMount=new Set},W=(a,b)=>(a&&b&&I(i,a,b),!ap(ag(),h)),X=(a,b,c)=>{let d={...k.mount?i:n(b)?h:A(a)?{[a]:b}:b};return B(a,r,d,c)},Y=b=>m(o(k.mount?i:h,b,a.shouldUnregister?o(h,b,[]):[])),Z=(a,b,c={})=>{let f=o(d,a),h=b;if(f){let j=f._f;j&&(j.disabled||I(i,a,ax(b,j)),h=aa&&aq(j.ref)&&g(b)?"":b,ar(j.ref)?[...j.ref.options].forEach(a=>a.selected=h.includes(a.value)):j.refs?e(j.ref)?j.refs.length>1?j.refs.forEach(a=>!a.disabled&&(a.checked=Array.isArray(h)?!!h.find(b=>b===a.value):h===a.value)):j.refs[0]&&(j.refs[0].checked=!!h):j.refs.forEach(a=>a.checked=a.value===h):P(j.ref)?j.ref.value="":(j.ref.value=h,j.ref.type||y.watch.next({name:a})))}(c.shouldDirty||c.shouldTouch)&&Q(a,h,c.shouldTouch,c.shouldDirty,!0),c.shouldValidate&&af(a)},_=(a,b,c)=>{for(let e in b){let g=b[e],h=`${a}.${e}`,i=o(d,h);!r.array.has(a)&&ao(g)&&(!i||i._f)||f(g)?Z(h,g,c):_(h,g,c)}},ac=(a,b,e={})=>{let f=o(d,a),j=r.array.has(a),k=ab(b);I(i,a,k),j?(y.array.next({name:a,values:i}),(w.isDirty||w.dirtyFields)&&e.shouldDirty&&(c.dirtyFields=aw(h,i),y.state.next({name:a,dirtyFields:c.dirtyFields,isDirty:W(a,k)}))):!f||f._f||g(k)?Z(a,k,e):_(a,k,e),M(a,r)&&y.state.next({}),y.watch.next({name:a})},ae=async a=>{let e=a.target,f=e.name,g=o(d,f);if(g){let h,k,l=e.type?ay(g._f):j(a),m=a.type===p.BLUR||a.type===p.FOCUS_OUT,n=!aB(g._f)&&!b.resolver&&!o(c.errors,f)&&!g._f.deps||aD(m,o(c.touchedFields,f),c.isSubmitted,D,z),q=M(f,r,m);I(i,f,l),m?(g._f.onBlur&&g._f.onBlur(a),s&&s(0)):g._f.onChange&&g._f.onChange(a);let t=Q(f,l,m,!1),w=!v(t)||q;if(m||y.watch.next({name:f,type:a.type}),n)return w&&y.state.next({name:f,...q?{}:t});if(!m&&q&&y.state.next({}),u[f]=(u[f],1),y.state.next({isValidating:!0}),b.resolver){let{errors:x}=await S([f]),A=aC(c.errors,d,f),B=aC(x,d,A.name||f);h=B.error,f=B.name,k=v(x)}else h=(await $(g,o(i,f),E,b.shouldUseNativeValidation))[f],k=await G(!0);g._f.deps&&af(g._f.deps),R(f,k,h,t)}},af=async(a,e={})=>{let f,g,h=x(a);if(y.state.next({isValidating:!0}),b.resolver){let i=await T(n(a)?a:h);f=v(i),g=a?!h.some(a=>o(i,a)):f}else a?((g=(await Promise.all(h.map(async a=>{let b=o(d,a);return await U(b&&b._f?{[a]:b}:b)}))).every(Boolean))||c.isValid)&&G():g=f=await U(d);return y.state.next({...!A(a)||w.isValid&&f!==c.isValid?{}:{name:a},...b.resolver?{isValid:f}:{},errors:c.errors,isValidating:!1}),e.shouldFocus&&!g&&J(d,a=>o(c.errors,a),a?h:r.mount),g},ag=a=>{let b={...h,...k.mount?i:{}};return n(a)?b:A(a)?o(b,a):a.map(a=>o(b,a))},ah=(a,b)=>({invalid:!!o((b||c).errors,a),isDirty:!!o((b||c).dirtyFields,a),isTouched:!!o((b||c).touchedFields,a),error:o((b||c).errors,a)}),ai=a=>{a?x(a).forEach(a=>ak(c.errors,a)):c.errors={},y.state.next({errors:c.errors})},aj=(a,b,e)=>{let f=(o(d,a,{_f:{}})._f||{}).ref;I(c.errors,a,{...b,ref:f}),y.state.next({name:a,errors:c.errors,isValid:!1}),e&&e.shouldFocus&&f&&f.focus&&f.focus()},al=(a,b)=>C(a)?y.watch.subscribe({next:c=>a(X(void 0,b),c)}):X(a,b,!0),am=(a,e={})=>{for(let f of a?x(a):r.mount)r.mount.delete(f),r.array.delete(f),o(d,f)&&(e.keepValue||(ak(d,f),ak(i,f)),e.keepError||ak(c.errors,f),e.keepDirty||ak(c.dirtyFields,f),e.keepTouched||ak(c.touchedFields,f),b.shouldUnregister||e.keepDefaultValue||ak(h,f));y.watch.next({}),y.state.next({...c,...e.keepDirty?{isDirty:W()}:{}}),e.keepIsValid||G()},au=(a,c={})=>{let e=o(d,a),f=O(c.disabled);return I(d,a,{_f:{...e&&e._f?e._f:{ref:{name:a}},name:a,mount:!0,...c}}),r.mount.add(a),e?f&&I(i,a,c.disabled?void 0:o(i,a,ay(e._f))):L(a,!0,c.value),{...f?{disabled:c.disabled}:{},...b.shouldUseNativeValidation?{required:!!c.required,min:aA(c.min),max:aA(c.max),minLength:aA(c.minLength),maxLength:aA(c.maxLength),pattern:aA(c.pattern)}:{},name:a,onChange:ae,onBlur:ae,ref(f){if(f){au(a,c),e=o(d,a);let g=n(f.value)&&f.querySelectorAll&&f.querySelectorAll("input,select,textarea")[0]||f,i=as(g),j=e._f.refs||[];(i?!j.find(a=>a===g):g!==e._f.ref)&&(I(d,a,{_f:{...e._f,...i?{refs:[...j.filter(at),g,...Array.isArray(o(h,a))?[{}]:[],],ref:{type:g.type,name:a}}:{ref:g}}}),L(a,!1,void 0,g))}else(e=o(d,a,{}))._f&&(e._f.mount=!1),(b.shouldUnregister||c.shouldUnregister)&&!(l(r.array,a)&&k.action)&&r.unMount.add(a)}}},av=(a,e)=>async f=>{f&&(f.preventDefault&&f.preventDefault(),f.persist&&f.persist());let g=!0,h=ab(i);try{if(b.resolver){let{errors:j,values:k}=await S();c.errors=j,h=k}else await U(d);v(c.errors)?(y.state.next({errors:{},isSubmitting:!0}),await a(h,f)):(e&&await e({...c.errors},f),b.shouldFocusError&&J(d,a=>o(c.errors,a),r.mount)),y.state.next({isSubmitting:!0})}catch(l){throw g=!1,l}finally{c.isSubmitted=!0,y.state.next({isSubmitted:!0,isSubmitting:!1,isSubmitSuccessful:v(c.errors)&&g,submitCount:c.submitCount+1,errors:c.errors})}},aG=(a,b={})=>{o(d,a)&&(n(b.defaultValue)?ac(a,o(h,a)):(ac(a,b.defaultValue),I(h,a,b.defaultValue)),b.keepTouched||ak(c.touchedFields,a),b.keepDirty||(ak(c.dirtyFields,a),c.isDirty=b.defaultValue?W(a,o(h,a)):W()),!b.keepError&&(ak(c.errors,a),w.isValid&&G()),y.state.next({...c}))},aH=(b,e={})=>{let f=b||h,g=ab(f),j=b&&!v(b)?g:h;if(e.keepDefaultValues||(h=f),!e.keepValues){if(e.keepDirtyValues)for(let l of r.mount)o(c.dirtyFields,l)?I(j,l,o(i,l)):ac(l,o(j,l));else{if(aa&&n(b))for(let m of r.mount){let p=o(d,m);if(p&&p._f){let q=Array.isArray(p._f.refs)?p._f.refs[0]:p._f.ref;try{if(aq(q)){q.closest("form").reset();break}}catch(s){}}}d={}}i=a.shouldUnregister?e.keepDefaultValues?ab(h):{}:g,y.array.next({values:j}),y.watch.next({values:j})}r={mount:new Set,unMount:new Set,array:new Set,watch:new Set,watchAll:!1,focus:""},k.mount=!w.isValid||!!e.keepIsValid,k.watch=!!a.shouldUnregister,y.state.next({submitCount:e.keepSubmitCount?c.submitCount:0,isDirty:e.keepDirty||e.keepDirtyValues?c.isDirty:!!(e.keepDefaultValues&&!ap(b,h)),isSubmitted:!!e.keepIsSubmitted&&c.isSubmitted,dirtyFields:e.keepDirty||e.keepDirtyValues?c.dirtyFields:e.keepDefaultValues&&b?aw(h,b):{},touchedFields:e.keepTouched?c.touchedFields:{},errors:e.keepErrors?c.errors:{},isSubmitting:!1,isSubmitSuccessful:!1})},aI=(a,b={})=>{let c=o(d,a)._f,e=c.refs?c.refs[0]:c.ref;e.focus(),b.shouldSelect&&e.select()};return{control:{register:au,unregister:am,getFieldState:ah,_executeSchema:S,_getWatch:X,_getDirty:W,_updateValid:G,_removeUnmounted:V,_updateFieldArray:H,_getFieldArray:Y,_subjects:y,_proxyFormState:w,get _fields(){return d},get _formValues(){return i},get _stateFlags(){return k},set _stateFlags(value){k=value},get _defaultValues(){return h},get _names(){return r},set _names(value){r=value},get _formState(){return c},set _formState(value){c=value},get _options(){return b},set _options(value){b={...b,...value}}},trigger:af,register:au,handleSubmit:av,watch:al,setValue:ac,getValues:ag,reset:aH,resetField:aG,clearErrors:ai,unregister:am,setError:aj,setFocus:aI,getFieldState:ah}}(a),formState:c};let i=b.current.control,k=d.useCallback(a=>{w(a,i._proxyFormState,!0)&&(i._formState={...i._formState,...a},h({...i._formState}))},[i]);return z({subject:i._subjects.state,callback:k}),d.useEffect(()=>{i._stateFlags.mount||(i._proxyFormState.isValid&&i._updateValid(),i._stateFlags.mount=!0),i._stateFlags.watch&&(i._stateFlags.watch=!1,i._subjects.state.next({})),i._removeUnmounted()}),b.current.formState=u(c,i._proxyFormState),b.current}}}])