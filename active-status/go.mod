module git.sr.ht/~athorp96/forest-ex/active-status

go 1.14

require (
	gioui.org v0.0.0-20200726090339-83673ecb203f
	git.sr.ht/~whereswaldon/colorpicker v0.0.0-20200801012301-b0b7a5822cd7
	git.sr.ht/~whereswaldon/forest-go v0.0.0-20200625210621-d3d4a318419f
	git.sr.ht/~whereswaldon/materials v0.0.0-20200801012148-3e241edb74da
	git.sr.ht/~whereswaldon/niotify v0.0.4-0.20200801012408-6296d10fa0f8
	git.sr.ht/~whereswaldon/sprout-go v0.0.0-20200517010141-a4188845a9a8
	git.wow.st/gmp/jni v0.0.0-20200709210836-4a3b173acb9f // indirect
	golang.org/x/crypto v0.0.0-20200728195943-123391ffb6de
	golang.org/x/exp v0.0.0-20200513190911-00229845015e
	golang.org/x/sys v0.0.0-20200728102440-3e129f6d46b1 // indirect
	golang.org/x/text v0.3.3 // indirect
)

replace golang.org/x/crypto => github.com/ProtonMail/crypto v0.0.0-20200605105621-11f6ee2dd602

replace git.sr.ht/~whereswaldon/forest-go => ../../forest-go
replace git.sr.ht/~athorp96/forest-ex/expiration => ../expiration
