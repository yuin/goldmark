- [ ] goldmark is fully compliant with the CommonMark. Before submitting issue, you **must** read [CommonMark spec](https://spec.commonmark.org/0.29/) and confirm your output is different from [CommonMark online demo](https://spec.commonmark.org/dingus/).
    - [ ] **Extensions(Autolink without `<` `>`, Table, etc) are not part of CommonMark spec.** You should confirm your output is different from other official renderers correspond with an extension.
- [ ] **goldmark is not dedicated for Hugo**. If you are Hugo user and your issue was raised by your experience in Hugo, **you should consider create issue at Hugo repository at first** .
- [ ] Before you make a feature request, **you should consider implement the new feature as an extension by yourself** . To keep goldmark itself simple, most new features should be implemented as an extension.

Please answer the following before submitting your issue:

1. What version of goldmark are you using? : 
2. What version of Go are you using? : 
3. What operating system and processor architecture are you using? :
4. What did you do? :
5. What did you expect to see? :
6. What did you see instead? :
7. Did you confirm your output is different from [CommonMark online demo](https://spec.commonmark.org/dingus/) or other official renderer correspond with an extension?:
8. (Feature request only): Why you can not implement it as an extension?:
    - You should avoid saying like "I'm not familiar with this project" "I'm not a Go programmer" as far as possible. This is an open source project and a library for Go programmers. I encourage you to strive to read source codes. 
    - I absolutely welcome questions that are difficult even if you read the source codes.
