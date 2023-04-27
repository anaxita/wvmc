package dal

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/anaxita/wvmc/internal/app"
	"github.com/anaxita/wvmc/internal/entity"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	user entity.User // already exists in the database
	cfg  *app.Config
	db   *sqlx.DB
	repo *UserRepository
}

func TestExampleTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}

func (s *UserRepositoryTestSuite) SetupSuite() {
	c, err := app.NewConfig("../../.env_test")
	s.Require().NoError(err)
	s.cfg = c

	db, err := app.NewSQLite3Client(c.DB)
	s.Require().NoError(err)
	s.db = db

	err = app.UpMigrations(db.DB, c.DB.Name, "../../migrations")
	s.Require().NoError(err)
	s.repo = NewUserRepository(db)
}

func (s *UserRepositoryTestSuite) TearDownSuite() {
	err := s.db.Close()
	s.Require().NoError(err)

	err = os.Remove(s.cfg.DB.Name)
	s.Require().NoError(err)
}

func (s *UserRepositoryTestSuite) TearDownTest() {
	_, err := s.db.Exec("DELETE FROM users")
	s.Require().NoError(err)
}

func (s *UserRepositoryTestSuite) TestCreateUser() {
	createdUser := generateUserCreate()

	ctx := context.Background()

	err := s.repo.Create(ctx, createdUser)
	s.Require().NoError(err)

	got, err := s.repo.FindByID(ctx, createdUser.ID)
	s.Require().NoError(err)

	expected := entity.User{
		ID:       createdUser.ID,
		Name:     createdUser.Name,
		Email:    createdUser.Email,
		Company:  createdUser.Company,
		Role:     createdUser.Role,
		Password: createdUser.Password,
	}

	s.Require().Equal(expected, got)
}

func (s *UserRepositoryTestSuite) TestCreateUserError() {
	createdUser := generateUserCreate()

	ctx := context.Background()

	err := s.repo.Create(ctx, createdUser)
	s.Require().NoError(err)

	// Try to create user with the same id and email
	err = s.repo.Create(ctx, createdUser)
	s.Require().Error(err)
}

func (s *UserRepositoryTestSuite) TestUpdateUser() {
	ctx := context.Background()
	createdUser := generateUserCreate()

	err := s.repo.Create(ctx, createdUser)
	s.Require().NoError(err)

	editUser := entity.UserEdit{
		Name:    "test_changed_name",
		Company: "test_changed_company",
		Role:    entity.RoleUser,
	}

	err = s.repo.Update(ctx, createdUser.ID, editUser)
	s.Require().NoError(err)

	got, err := s.repo.FindByID(ctx, createdUser.ID)
	s.Require().NoError(err)

	expected := entity.User{
		ID:       createdUser.ID,
		Email:    createdUser.Email,
		Password: createdUser.Password,
		Name:     editUser.Name,
		Company:  editUser.Company,
		Role:     editUser.Role,
	}

	s.Require().Equal(expected, got)
}

func (s *UserRepositoryTestSuite) TestUpdateUserError() {
	ctx := context.Background()
	createdUser := generateUserCreate()

	err := s.repo.Create(ctx, createdUser)
	s.Require().NoError(err)

	editUser := entity.UserEdit{
		Name:    "test_changed_name",
		Company: "test_changed_company",
		Role:    entity.RoleUser,
	}

	// Try to update user with non-existent id
	err = s.repo.Update(ctx, uuid.New(), editUser)
	s.Require().Error(err)
}

func (s *UserRepositoryTestSuite) TestDeleteUser() {
	ctx := context.Background()
	createdUser := generateUserCreate()

	err := s.repo.Create(ctx, createdUser)
	s.Require().NoError(err)

	_, err = s.repo.FindByID(ctx, createdUser.ID)
	s.Require().NoError(err)

	err = s.repo.Delete(ctx, createdUser.ID)
	s.Require().NoError(err)

	_, err = s.repo.FindByID(ctx, createdUser.ID)
	s.Require().Error(err)
}

func (s *UserRepositoryTestSuite) TestDeleteUserError() {
	ctx := context.Background()
	err := s.repo.Delete(ctx, uuid.New())
	s.Require().Error(err)
}

func (s *UserRepositoryTestSuite) TestFindByID() {
	ctx := context.Background()
	createdUser := generateUserCreate()

	err := s.repo.Create(ctx, createdUser)
	s.Require().NoError(err)

	got, err := s.repo.FindByID(ctx, createdUser.ID)
	s.Require().NoError(err)

	expected := entity.User{
		ID:       createdUser.ID,
		Email:    createdUser.Email,
		Password: createdUser.Password,
		Name:     createdUser.Name,
		Company:  createdUser.Company,
		Role:     createdUser.Role,
	}

	s.Require().Equal(expected, got)
}

func (s *UserRepositoryTestSuite) TestFindByIDError() {
	ctx := context.Background()
	_, err := s.repo.FindByID(ctx, uuid.New())
	s.Require().Error(err)
}

func (s *UserRepositoryTestSuite) TestFindByEmail() {
	ctx := context.Background()
	createdUser := generateUserCreate()

	err := s.repo.Create(ctx, createdUser)
	s.Require().NoError(err)

	got, err := s.repo.FindByEmail(ctx, createdUser.Email)
	s.Require().NoError(err)

	expected := entity.User{
		ID:       createdUser.ID,
		Email:    createdUser.Email,
		Password: createdUser.Password,
		Name:     createdUser.Name,
		Company:  createdUser.Company,
		Role:     createdUser.Role,
	}

	s.Require().Equal(expected, got)
}

func (s *UserRepositoryTestSuite) TestFindByEmailError() {
	ctx := context.Background()
	_, err := s.repo.FindByEmail(ctx, "test_email")
	s.Require().Error(err)
}

func (s *UserRepositoryTestSuite) TestUsers() {
	ctx := context.Background()
	users := generateUsersCreate(10)

	for _, user := range users {
		err := s.repo.Create(ctx, user)
		s.Require().NoError(err)
	}

	got, err := s.repo.Users(ctx)
	s.Require().NoError(err)

	s.Require().Len(got, 10)
}

type userOpt func(user *entity.UserCreate)

// withUserRole is a userOpt that sets the role to user.
func withUserRole() userOpt {
	return func(user *entity.UserCreate) {
		user.Role = entity.RoleUser
	}
}

func generateUserCreate(opts ...userOpt) entity.UserCreate {
	user := entity.UserCreate{
		ID:       uuid.New(),
		Name:     "test_name_",
		Email:    "test_email" + uuid.New().String(),
		Company:  "test_company",
		Role:     entity.RoleAdmin,
		Password: "test_password",
	}

	for _, opt := range opts {
		opt(&user)
	}

	return user
}

func generateUsersCreate(count int) []entity.UserCreate {
	users := make([]entity.UserCreate, count)
	for i := 0; i < count; i++ {
		user := generateUserCreate()
		user.Email = fmt.Sprintf("%s_%d", user.Email, i)
		user.Name = fmt.Sprintf("%s_%d", user.Name, i)
		user.Company = fmt.Sprintf("%s_%d", user.Company, i)

		users[i] = user
	}
	return users
}
