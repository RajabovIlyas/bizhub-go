(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[701],{30663:function(a,b,c){(window.__NEXT_P=window.__NEXT_P||[]).push(["/panel/feedbacks",function(){return c(98811)}])},92456:function(a,b,c){"use strict";var d=c(26042),e=c(69396),f=c(99534),g=c(85893),h=c(94184),i=c.n(h);c(67294);var j=function(a){var b=a.children,c=a.className,h=(0,f.Z)(a,["children","className"]);return(0,g.jsx)("div",(0,e.Z)((0,d.Z)({className:i()("w-full flex flex-col h-full",c)},h),{children:b}))};b.Z=j},7641:function(a,b,c){"use strict";c.d(b,{V2:function(){return l},ZX:function(){return n},dR:function(){return m}});var d=c(26042),e=c(69396),f=c(99534),g=c(85893),h=c(94184),i=c.n(h),j=c(67294),k=c(14212),l=function(a){var b=a.children;return(0,g.jsxs)("div",{className:"flex items-center justify-between",children:[1===j.Children.count(b)&&(0,g.jsx)("div",{}),j.Children.map(b,function(a){return a})]})},m=function(a){var b=a.children,c=a.paddingYDisabled,d=a.style;return(0,g.jsx)("div",{style:d,className:i()("flex items-center space-x-7 px-7",!c&&"py-7"),children:b})},n=function(a){var b=a.children,c=a.icon,h=a.color,j=void 0===h?"primary":h,l=(0,f.Z)(a,["children","icon","color"]);return(0,g.jsxs)("button",(0,e.Z)((0,d.Z)({},l),{className:i()("flex items-center space-x-2 py-2 px-3.5 rounded-main border","primary"===j&&"border-primary bg-primary","outline"===j&&"bg-white"),children:[c&&(0,g.jsx)(k.Z,{name:c,color:"primary"===j?"white":"black",size:15}),(0,g.jsx)("span",{className:i()("text-[15px] font-medium","primary"===j&&"text-white","outline"===j&&"text-black"),children:b})]}))}},12906:function(a,b,c){"use strict";var d=c(85893),e=c(67294),f=function(a){var b=a.id,c=a.options,f=a.onChange,g=a.value,h=void 0===g?0:g,i=a.disabled,j=(0,e.useState)(null),k=j[0],l=j[1],m=function(a){return function(){if(!i){l(a);var c=document.getElementById("ojo-animated-select-group-".concat(b,"-container")),d=document.getElementById("ojo-animated-select-group-".concat(b,"-selected")),e=document.getElementById("ojo-animated-select-group-".concat(b,"-item-").concat(a));if(e&&d&&c){var g,h=g=e.getBoundingClientRect().left-c.getBoundingClientRect().left;d.style.width=(null==e?void 0:e.offsetWidth)+"px",d.style.height=((null==e?void 0:e.offsetHeight)>(null==c?void 0:c.offsetHeight)-10?null==e?void 0:e.offsetHeight:(null==c?void 0:c.offsetHeight)-10)+"px",d.style.transform="translateX(".concat(h,"px)")}f&&f(a)}}};return(0,e.useEffect)(function(){setTimeout(function(){null===k&&m(h)()},1)},[]),(0,e.useEffect)(function(){if(void 0!==h){var a;(a=h)<c.length&&a>=0&&m(h)()}},[h]),(0,d.jsxs)("div",{id:"ojo-animated-select-group-".concat(b,"-container"),className:"rounded-main relative bg-[#E5E5E5] flex items-center p-[5px]",children:[(0,d.jsx)("div",{className:"bg-white transition-all z-[0] absolute left-0 duration-[200ms] rounded-[7px]",style:{boxShadow:"0px 0px 4px 0px rgba(0, 0, 0, 0.15)"},id:"ojo-animated-select-group-".concat(b,"-selected")}),c.map(function(a,c){return(0,d.jsx)("button",{disabled:i,id:"ojo-animated-select-group-".concat(b,"-item-").concat(c),onClick:k===c?void 0:m(c),className:"px-[11px] transition-all duration-[200ms] outline-none focus:text-primary ".concat(k===c?"text-black":"text-gray-500"," z-[1] py-[5px] rounded-[7px]"),children:a},"ro-index-".concat(b,"-").concat(c))})]})};b.Z=f},38560:function(a,b,c){"use strict";c.d(b,{E:function(){return j},d:function(){return i}});var d=c(26042),e=c(69396),f=c(85893),g=c(70211),h=c(14212),i=function(a){return(0,f.jsx)(g.ZP,(0,d.Z)({menuPlacement:"bottom",components:{Option:function(a){return(0,f.jsx)("div",(0,e.Z)((0,d.Z)({style:(0,d.Z)({},a.isSelected?{color:"black"}:{color:"#AAAAAA"},a.isFocused?{backgroundColor:"rgb(170 170 170 / 0.05)"}:{backgroundColor:"transparent"}),className:"cursor-pointer whitespace-nowrap flex hover:bg-[#AAAAAA]/5 items-center py-[7.5px] px-[15px]"},a.innerProps),{children:(0,f.jsx)("span",{className:"whitespace-wrap",children:a.label})}))},IndicatorsContainer:function(){return(0,f.jsx)("div",{className:"pr-[15px]",children:(0,f.jsx)(h.Z,{name:"chevronDown",className:"stroke-black fill-transparent",size:15})})}},noOptionsMessage:function(a){var b=a.inputValue;return(0,f.jsxs)("span",{children:[(0,f.jsx)("span",{className:"text-red-500",children:b})," not found"]})},styles:{container:function(a,b){return(0,e.Z)((0,d.Z)({},a),{padding:0,margin:0,borderColor:"blue",":focus":{boxShadow:"none"}})},control:function(a,b){return(0,e.Z)((0,d.Z)({},a),{borderRadius:"10px",padding:"0.5rem 0px",boxShadow:"none",":focus":{boxShadow:"none",borderColor:"#E5E5E5"},borderColor:"#E5E5E5",":hover":{borderColor:"#E5E5E5",boxShadow:"none"},":active":{boxShadow:"none"}})},input:function(a,b){return(0,e.Z)((0,d.Z)({},a),{padding:0,margin:0})},indicatorSeparator:function(a,b){return(0,e.Z)((0,d.Z)({},a),{display:"none",width:0,height:0,padding:0,margin:0})},valueContainer:function(a,b){return(0,e.Z)((0,d.Z)({},a),{paddingTop:"0px",paddingBottom:"0px",paddingLeft:"15px",fontSize:"15px"})},menu:function(a,b){return(0,e.Z)((0,d.Z)({},a),{minWidth:"100px",width:"auto",borderRadius:"10px",right:0,boxShadow:"0px 0px 6px 0px rgba(0,0,0,0.15)"})}}},a))},j=function(a){return(0,f.jsx)(g.ZP,(0,d.Z)({components:{Option:function(a){return(0,f.jsx)("div",(0,e.Z)((0,d.Z)({style:(0,d.Z)({},a.isSelected?{color:"black"}:{color:"#AAAAAA"},a.isFocused?{backgroundColor:"rgb(170 170 170 / 0.05)"}:{backgroundColor:"transparent"}),className:"cursor-pointer flex hover:bg-[#AAAAAA]/5 items-center py-[7.5px] px-[15px]"},a.innerProps),{children:(0,f.jsx)("span",{children:a.label})}))},IndicatorsContainer:function(){return(0,f.jsx)(f.Fragment,{})}},noOptionsMessage:function(a){var b=a.inputValue;return(0,f.jsxs)("span",{children:[(0,f.jsx)("span",{className:"text-red-500",children:b})," not found"]})},styles:{container:function(a,b){return(0,e.Z)((0,d.Z)({},a),{padding:0,margin:0,borderColor:"blue",":focus":{boxShadow:"none"}})},control:function(a,b){return(0,e.Z)((0,d.Z)({},a),{borderRadius:"10px",boxShadow:"none",":focus":{boxShadow:"none",borderColor:"#E5E5E5"},":hover":{borderColor:"#E5E5E5",boxShadow:"none"},":active":{boxShadow:"none"},padding:"8px 15px",backgroundColor:"#F7F8FA",borderColor:"#E5E5E5",cursor:"pointer"})},input:function(a,b){return(0,e.Z)((0,d.Z)({},a),{padding:0,margin:0})},indicatorSeparator:function(a,b){return(0,e.Z)((0,d.Z)({},a),{display:"none",width:0,height:0,padding:0,margin:0})},valueContainer:function(a,b){return(0,e.Z)((0,d.Z)({},a),{padding:0})},menu:function(a,b){return(0,e.Z)((0,d.Z)({},a),{minWidth:"190px",borderRadius:"10px",right:0,overflow:"hidden",boxShadow:"0px 0px 6px 0px rgba(0,0,0,0.15)"})}}},a))}},98811:function(a,b,c){"use strict";c.r(b),c.d(b,{default:function(){return z}});var d=c(47568),e=c(14924),f=c(10797),g=c(34051),h=c.n(g),i=c(85893),j=c(9669),k=c.n(j),l=c(67294),m=c(82668),n=c(14212),o=c(92456),p=c(7641),q=c(12906),r=c(38560),s=c(32130),t=c(67971),u=c(65791),v=c(75758),w=c(94989),x=function(){var a,b,c,d,e=(0,l.useState)(!1),f=e[0],g=e[1],h=(0,l.useState)(null),j=h[0],k=h[1],m=function(a){k(a),g(!0)},o=function(){g(!1),k(null)};(0,l.useEffect)(function(){return v.eventManager.on("feedbacks:info:show",m),v.eventManager.on("feedbacks:info:hide",o),function(){v.eventManager.off("feedbacks:info:show",m),v.eventManager.off("feedbacks:info:hide",o)}},[]);var p=(0,w.Y)((null==j?void 0:j.created_at)?new Date(j.created_at):new Date);return(0,i.jsx)(u.Z,{onClose:o,className:"flex flex-col flex-grow",visible:f,title:"Notification Info",customTitle:(0,i.jsxs)("div",{className:"flex items-center space-x-[15px]",children:[(0,i.jsx)("div",{className:"overflow-hidden w-[40px] h-[40px] rounded-full",children:(0,i.jsx)("img",{alt:null==j?void 0:j.seller.name,className:"w-full h-full",src:"/cdn/".concat(null==j?void 0:j.seller.logo)})}),(0,i.jsxs)("div",{className:"flex flex-col items-start",children:[(0,i.jsx)("h4",{className:"font-[600] text-[16px]",children:null==j?void 0:null===(a=j.seller)|| void 0===a?void 0:a.name}),(0,i.jsx)("span",{className:"text-[#9797BE] font-[400] text-[12px]",children:null==j?void 0:null===(b=j.seller)|| void 0===b?void 0:null===(c=b.city)|| void 0===c?void 0:c.name})]}),(null==j?void 0:null===(d=j.seller)|| void 0===d?void 0:d.type)==="manufacturer"&&(0,i.jsx)(n.Z,{name:"verified",size:22})]}),children:(0,i.jsxs)("div",{className:"flex flex-col flex-grow space-y-[15px]",children:[(0,i.jsxs)("span",{className:"font-[400] text-[12px] text-[#777777]",children:[p.getDay()," ",p.getMonthAsString(),","," ",p.getYear(),"y. / ",p.getHoursAsString(),":",p.getMinutesAsString()]}),(0,i.jsx)("p",{className:"text-[16px] font-[500]",children:null==j?void 0:j.text})]})})},y=x,z=function(){var a,b,c,g=(0,l.useState)("all"),j=g[0],u=g[1],x=(0,l.useState)({value:"all",label:"All"}),z=x[0],A=x[1],B=(0,l.useState)(!1),C=B[0],D=B[1],E=(0,l.useState)(!1),F=(E[0],E[1]),G=(0,l.useState)(0),H=G[0],I=G[1],J=(0,l.useState)([]),K=J[0],L=J[1],M=(a=(0,d.Z)(h().mark(function a(b,c,d){var g,i;return h().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return D(!0),a.prev=1,g=!0===b?0:H,a.next=5,k().get("/api/v1/feedbacks",{params:{page:g,limit:10,sort:c,filter:d},headers:(0,e.Z)({},s.G.AddAccessTokenToHeader,!0)});case 5:!0===(i=a.sent).data.isSuccess&&(F(i.data.result.length<10),I(g+1),L(function(a){return!0===b?i.data.result:(0,f.Z)(a).concat((0,f.Z)(i.data.result))})),D(!1),a.next=14;break;case 10:a.prev=10,a.t0=a.catch(1),console.log("[loadFeedbacks] - error -",a.t0),D(!1);case 14:case"end":return a.stop()}},a,null,[[1,10]])})),function(b,c,d){return a.apply(this,arguments)});return(0,l.useEffect)(function(){M()},[]),(0,i.jsxs)(o.Z,{className:"relative",children:[(0,i.jsxs)(p.V2,{children:[(0,i.jsx)(p.dR,{children:(0,i.jsx)(q.Z,{id:"feedbacks-sort-anim",value:"all"===j?0:"unread"===j?1:2,onChange:(b=(0,d.Z)(h().mark(function a(b){var c;return h().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return u(c=0===b?"all":1===b?"unread":"read"),a.next=4,M(!0,c,z.value);case 4:return a.abrupt("return",a.sent);case 5:case"end":return a.stop()}},a)})),function(a){return b.apply(this,arguments)}),options:["All","Unread","Read"]})},"left"),(0,i.jsx)(p.dR,{children:(0,i.jsx)(r.d,{value:z,onChange:(c=(0,d.Z)(h().mark(function a(b,c){var d;return h().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:if(b.value!==z.value){a.next=2;break}return a.abrupt("return");case 2:return A(b),console.log("filter =>",b),a.next=7,M(!0,null!=j?j:"all",null!==(d=b.value)&& void 0!==d?d:"all");case 7:return a.abrupt("return",a.sent);case 8:case"end":return a.stop()}},a)})),function(a,b){return c.apply(this,arguments)}),options:[{value:"all",label:"All"},{value:"today",label:"Today"},{value:"this_week",label:"This Week"},{value:"this_month",label:"This Month"},]})},"right")]}),C&&(0,i.jsx)("div",{className:"flex absolute w-full h-full z-[10] flex-grow bg-[rgba(255,255,255,0.4)] items-center justify-center py-10",children:(0,i.jsx)(m.Z,{color:t.Colors.primary,width:"70px",loading:C})}),(0,i.jsx)("div",{className:"grid grid-cols-1 gap-4 px-7",children:K.map(function(a,b){var c=(0,w.Y)(new Date(a.created_at));return(0,i.jsxs)("div",{onClick:function(){v.eventManager.emit("feedbacks:info:show",a)},className:"shadow-main cursor-pointer justify-between relative lg:flex-row flex-col bg-white rounded-main flex items-start lg:items-center py-3 sm:py-[7px] px-6",children:[(0,i.jsxs)("div",{className:"flex items-center lg:my-0 my-2",children:[!a.is_read&&(0,i.jsx)("div",{className:"w-3.5 h-3.5 min-w-3.5 min-h-3.5 lg:mr-[20px] mr-0 rounded-full lg:relative absolute lg:top-0 lg:left-0 -top-1.5 -left-1.5 bg-secondary"}),(0,i.jsxs)("div",{className:"flex items-center space-x-[20px]",children:[(0,i.jsx)("div",{className:"w-[40px] border h-[40px] rounded-full overflow-hidden",children:(0,i.jsx)("img",{alt:a.seller.name,className:"w-full h-full",src:"/cdn/".concat(a.seller.logo)})}),(0,i.jsx)("h1",{className:"text-[16px] whitespace-nowrap",children:a.seller.name}),(0,i.jsx)("h3",{className:"text-[#9797BE] text-[14px]",children:a.seller.city.name}),"manufacturer"===a.seller.type&&(0,i.jsx)(n.Z,{name:"verified",size:22})]})]}),(0,i.jsxs)("div",{className:"flex items-center space-x-4 ml-0 lg:ml-[20px] overflow-hidden flex-grow max-w-[100%]",children:[(0,i.jsx)("h5",{className:"text-[15px] flex-grow truncate font-normal text-[#6E5AD1]",children:a.text}),(0,i.jsxs)("span",{className:"text-[#777777] text-[14px]",children:[c.getHoursAsString(),":",c.getMinutesAsString()]})]})]},"feedback-index-".concat(b))})}),(0,i.jsx)(y,{})]})}}},function(a){a.O(0,[211,774,888,179],function(){var b;return a(a.s=30663)}),_N_E=a.O()}])