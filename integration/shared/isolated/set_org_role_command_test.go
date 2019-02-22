package isolated

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"

	"code.cloudfoundry.org/cli/integration/helpers"
)

var _ = Describe("set-org-role command", func() {
	Describe("help text and argument validation", func() {
		When("-h is passed", func() {
			It("prints the help text", func() {
				session := helpers.CF("set-org-role", "-h")
				Eventually(session).Should(Say(`NAME:`))
				Eventually(session).Should(Say(`\s+set-org-role - Assign an org role to a user`))
				Eventually(session).Should(Say(`USAGE:`))
				Eventually(session).Should(Say(`\s+cf set-org-role USERNAME ORG ROLE`))
				Eventually(session).Should(Say(`ROLES:`))
				Eventually(session).Should(Say(`\s+'OrgManager' - Invite and manage users, select and change plans, and set spending limits`))
				Eventually(session).Should(Say(`\s+'BillingManager' - Create and manage the billing account and payment info`))
				Eventually(session).Should(Say(`\s+'OrgAuditor' - Read-only access to org info and reports`))
				Eventually(session).Should(Say(`SEE ALSO:`))
				Eventually(session).Should(Say(`\s+org-users, set-space-role`))
				Eventually(session).Should(Exit(0))
			})
		})

		When("not enough arguments are provided", func() {
			It("prints an error and help text", func() {
				session := helpers.CF("set-org-role", "foo", "bar")
				Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `ROLE` was not provided"))
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("   set-org-role - Assign an org role to a user"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("   cf set-org-role USERNAME ORG ROLE"))
				Eventually(session).Should(Say("ROLES:"))
				Eventually(session).Should(Say("   'OrgManager' - Invite and manage users, select and change plans, and set spending limits"))
				Eventually(session).Should(Say("   'BillingManager' - Create and manage the billing account and payment info"))
				Eventually(session).Should(Say("   'OrgAuditor' - Read-only access to org info and reports"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("   org-users, set-space-role"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("too many arguments are provided", func() {
			It("prints an error and help text", func() {
				session := helpers.CF("set-org-role", "some-user", "some-org", "OrgManager", "some-extra-argument")
				Eventually(session).Should(Say(`Incorrect Usage. Requires USERNAME, ORG, ROLE as arguments`))
				Eventually(session).Should(Say(`NAME:`))
				Eventually(session).Should(Say(`\s+set-org-role - Assign an org role to a user`))
				Eventually(session).Should(Say(`USAGE:`))
				Eventually(session).Should(Say(`\s+cf set-org-role USERNAME ORG ROLE`))
				Eventually(session).Should(Say(`ROLES:`))
				Eventually(session).Should(Say(`\s+'OrgManager' - Invite and manage users, select and change plans, and set spending limits`))
				Eventually(session).Should(Say(`\s+'BillingManager' - Create and manage the billing account and payment info`))
				Eventually(session).Should(Say(`\s+'OrgAuditor' - Read-only access to org info and reports`))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	When("the user is logged in", func() {
		var orgName string
		var username string

		BeforeEach(func() {
			helpers.LoginCF()
			orgName = ReadOnlyOrg
			username, _ = helpers.CreateUser()
		})

		When("the org and user both exist", func() {
			It("sets the org role for the user", func() {
				session := helpers.CF("set-org-role", username, orgName, "OrgAuditor")
				Eventually(session).Should(Say("Assigning role OrgAuditor to user %s in org %s as admin...", username, orgName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
			})

			When("the logged in user has insufficient permissions", func() {
				BeforeEach(func() {
					helpers.SwitchToOrgRole(orgName, "OrgAuditor")
				})

				It("prints out the error message from CC API and exits 1", func() {
					session := helpers.CF("set-org-role", username, orgName, "OrgAuditor")
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Say("Server error, status code: 403, error code: 10003, message: You are not authorized to perform the requested action"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the user already has the desired role", func() {
				BeforeEach(func() {
					session := helpers.CF("set-org-role", username, orgName, "OrgManager")
					Eventually(session).Should(Say("Assigning role OrgManager to user %s in org %s as admin...", username, orgName))
					Eventually(session).Should(Exit(0))
				})

				It("is idempotent", func() {
					session := helpers.CF("set-org-role", username, orgName, "OrgManager")
					Eventually(session).Should(Say("Assigning role OrgManager to user %s in org %s as admin...", username, orgName))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the specified role is invalid", func() {
				It("prints a useful error, prints help text, and exits 1", func() {
					session := helpers.CF("set-org-role", username, orgName, "NotARealRole")
					Eventually(session.Err).Should(Say(`Incorrect Usage: ROLE must be "OrgManager", "BillingManager" and "OrgAuditor"`))
					Eventually(session).Should(Say(`NAME:`))
					Eventually(session).Should(Say(`\s+set-org-role - Assign an org role to a user`))
					Eventually(session).Should(Say(`USAGE:`))
					Eventually(session).Should(Say(`\s+set-org-role USERNAME ORG ROLE`))
					Eventually(session).Should(Say(`ROLES:`))
					Eventually(session).Should(Say(`\s+'OrgManager' - Invite and manage users, select and change plans, and set spending limits`))
					Eventually(session).Should(Say(`\s+'BillingManager' - Create and manage the billing account and payment info`))
					Eventually(session).Should(Say(`\s+'OrgAuditor' - Read-only access to org info and reports`))
					Eventually(session).Should(Say(`SEE ALSO:`))
					Eventually(session).Should(Say(`\s+org-users, set-space-role`))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("the org does not exist", func() {
			It("prints an appropriate error and exits 1", func() {
				session := helpers.CF("set-org-role", username, "not-exists", "OrgAuditor")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Say("Organization not-exists not found"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the user does not exist", func() {
			It("prints an appropriate error and exits 1", func() {
				session := helpers.CF("set-org-role", "not-exists", orgName, "OrgAuditor")
				Eventually(session).Should(Say("Assigning role OrgAuditor to user not-exists in org %s as admin...", orgName))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Say("Server error, status code: 404, error code: 20003, message: The user could not be found: not-exists"))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	When("the user is not logged in", func() {
		BeforeEach(func() {
			helpers.LogoutCF()
		})

		It("reports that the user is not logged in", func() {
			session := helpers.CF("set-org-role", "some-user", "some-org", "BillingManager")
			Eventually(session).Should(Say("FAILED"))
			Eventually(session).Should(Say("Not logged in. Use 'cf login' to log in."))
			Eventually(session).Should(Exit(1))
		})
	})
})
